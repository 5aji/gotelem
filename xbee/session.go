/*
Package xbee provides communication and configuration of Digi XBee products

(and other Digi products that are similar such as the XLR Pro). It provides
a net.Conn-like interface as well as AT commands for configuration. The most
common usage of the package is with a Session, which provides
*/
package xbee

import (
	"bufio"
	"fmt"
	"io"
	"sync"

	"golang.org/x/exp/slog"
)

// TODO: implement net.Conn for Session/Conn. We are missing LocalAddr, RemoteAddr,
// and Deadline related methods.

// Session represents a connection to a locally-attached XBee. The connection can be through
// serial/USB or TCP/IP depending on what is supported by the device.
type Session struct {
	ioDev io.ReadWriteCloser
	ct    connTrack
	log   slog.Logger
	// this buffer is used for storing data that must be read at some point.
	rxBuf *bufio.ReadWriter

	writeLock sync.Mutex // prevents multiple writers from accessing the port at once.

	// conns contain a map of addresses to connections. This means that there
	// can only be one direct connection to a device. This is pretty reasonable IMO.
	// but needs to be documented very clearly.
	conns map[uint64]*Conn
}

// NewSession takes an IO device and a logger and returns a new XBee session.
func NewSession(dev io.ReadWriteCloser, baseLog *slog.Logger) (*Session, error) {
	sess := &Session{
		ioDev: dev,
		log:   *baseLog,
		ct:    *NewConnTrack(),
	}

	// setup io readwriter with a pipe.
	rd, wr := io.Pipe()
	// this is for reading data *only* - writes are different! it's a
	// readWriter because the goroutine that runs scan continuously (and handles all other packets)
	// will write to the buffer when new Rx packets come in, and we can read out from application code.
	sess.rxBuf = bufio.NewReadWriter(bufio.NewReader(rd), bufio.NewWriter(wr))

	// start the rx handler in the background. we close it later by closing the serial port.

	go sess.rxHandler()

	return sess, nil
}

// before we can write `Read(p []byte)` we have to have a goroutine that takes the input from
// the serial port and parses it out - if it's data, we push the data to a buffer for
// the application to read the bytes on its own.
//
// if it's a different kind of packet, we do custom functionality (free the conntrack, update
// local status, etc)
func (sess *Session) rxHandler() {
	// we wrap the serial port read line in a bufio.scanner using our custom split function.
	scan := bufio.NewScanner(sess.ioDev)
	scan.Split(xbeeFrameSplit)

	// scan.Scan() will return false when there's EOF, i.e the io device is closed.
	// this is activated by sess.Close()
	for scan.Scan() {
		data, err := parseFrame(scan.Bytes())
		if err != nil {
			sess.log.Warn("error parsing frame", "error", err, "data", data)
			continue
		}

		switch XBeeCmd(data[0]) {
		case RxPktType:
			// we parse the data, and push it to the rx buffer.
			//TODO: if we have multiple remotes on the network, we need to track them here.
			frame, err := ParseRxFrame(data)
			if err != nil {
				sess.log.Warn("error parsing rx packet", "error", err, "data", data)
				break //continue?
			}
			// take the data and write it to our internal rx packet buffer.
			_, err = sess.rxBuf.Write(frame.Payload)
			if err != nil {
				sess.log.Warn("error writing data", "error", err, "payload", frame.Payload)
			}

		case TxStatusType, ATCmdResponseType, RemoteCmdRespType:
			// we hand the frame back via the channel. we directly find the ID since it's always
			// the second byte.
			idx := data[1]

			err := sess.ct.ClearMark(idx, data)
			if err != nil {
				// we got a rogue packet lol
				sess.log.Warn("rogue frame ID", "id", idx, "error", err)
			}

		default:
			// we don't know what to do with it.
			sess.log.Info("unhandled packet type", "type", data[0], "id", data[1])

		}

	}
	// if we get here, the serial port has closed. this is fine.
	sess.log.Debug("closing rx handler", "err", scan.Err())
}

// This implements io.Reader for the UART Session.
func (sess *Session) Read(p []byte) (int, error) {
	// Since we have an rx buffer, we just read from that and return the results.
	return sess.rxBuf.Read(p)
}

// Write sends a message to all XBees listening on the network. To send a message to a specific
// XBee, use Dial() to get a Conn
func (sess *Session) Write(p []byte) (int, error) {

	return sess.writeAddr(p, 0xFFFF)

}

func (sess *Session) writeAddr(p []byte, dest uint64) (n int, err error) {

	idx, ch, err := sess.ct.GetMark()
	if err != nil {
		return
	}

	wf := &TxFrame{
		Id:          idx,
		Destination: dest,
		Payload:     p,
	}

	sess.writeLock.Lock()
	n, err = writeXBeeFrame(sess.ioDev, wf.Bytes())
	sess.writeLock.Unlock()
	if err != nil {
		return
	}
	n = n - 4

	// finally, wait for the channel we got to return. this means that
	// the matching response frame was received, so we can parse it.
	// TODO: add timeout.
	responseFrame := <-ch

	// this is a tx status frame bytes, so lets parse it out.
	status, err := ParseTxStatusFrame(responseFrame)
	if err != nil {
		return
	}

	if status.Status != 0 {
		err = fmt.Errorf("tx failed 0x%x", status.Status)
	}

	return
}

// sends a local AT command. If `queued` is true, the command is not immediately applied;
// instead, an AC command must be set to apply the queued changes. `queued` does not
// affect query-type commands, which always return right away.
// the AT command is an interface.
func (sess *Session) ATCommand(cmd [2]rune, data []byte, queued bool) ([]byte, error) {
	// we must encode the command, and then create the actual packet.
	// then we send the packet, and wait for the response
	// TODO: how to handle multiple-response-packet AT commands?
	// (mainly Node Discovery ND)

	// get a mark for the frame

	idx, ch, err := sess.ct.GetMark()
	if err != nil {
		return nil, err
	}
	rawData := encodeATCommand(cmd, data, idx, queued)

	sess.writeLock.Lock()
	_, err = writeXBeeFrame(sess.ioDev, rawData)
	sess.writeLock.Unlock()

	if err != nil {
		return nil, fmt.Errorf("error writing xbee frame: %w", err)
	}

	// TODO: add timeout.

	resp, err := ParseATCmdResponse(<-ch)
	if err != nil {
		return nil, err
	}

	if resp.Status != 0 {
		// sinec ATCmdStatus is a stringer thanks to the generator
		return resp.Data, fmt.Errorf("AT command failed: %v", resp.Status)
	}

	return resp.Data, nil

}

// Does this need to exist?
func (sess *Session) GetStatus() {
	panic("TODO: implement")
}

// Implement the io.Closer.
func (sess *Session) Close() error {
	return sess.ioDev.Close()
}

// Conn is a connection to a specific remote XBee. Conn allows for the user to
// contact one Xbee for point-to-point communications. This enables ACK packets
// for reliable transmission.
type Conn struct {
	parent *Session
	log    slog.Logger
	addr   uint64
}

func (c *Conn) Write(p []byte) (int, error) {
	return c.parent.writeAddr(p, c.addr)
}

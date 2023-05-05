package xbee

import (
	"bufio"
	"fmt"
	"io"
	"sync"

	"go.bug.st/serial"
	"golang.org/x/exp/slog"
)

// todo: make transport-agnostic (serial port or TCP/IP)

// A session is a simple way to manage an xbee device.
// it provides io.Reader and io.Writer, as well as some extra functions to handle
// custom Xbee frames.
type Session interface {
	io.ReadWriteCloser
	GetStatus() // todo: figure out signature for this

	// Dial takes an address and allows direct communication with that
	// device, without using broadcast.
	Dial(addr uint64) io.ReadWriteCloser
	// AT command related functions - query, set on local, query, set on remote.

	ATCommand(cmd ATCmd, queued bool) (resp ATCmd, err error)
	RemoteATCommand(cmd ATCmd, addr uint64) (resp ATCmd, err error)
}

type SerialSession struct {
	port serial.Port
	ct   connTrack
	slog.Logger
	// todo: add queuing structures here for reliable transport and tracking.
	// this buffer is used for storing data that must be read at some point.
	rxBuf     *bufio.ReadWriter
	writeLock sync.Mutex // prevents multiple writers from accessing the port at once.
}

func NewSerialXBee(portName string, mode *serial.Mode) (*SerialSession, error) {
	// make the session with the port/mode given, and set up the conntrack.
	sess := &SerialSession{}

	port, err := serial.Open(portName, mode)
	if err != nil {
		return sess, err
	}
	sess.port = port

	sess.ct = connTrack{}

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
func (sess *SerialSession) rxHandler() {
	// we wrap the serial port read line in a bufio.scanner using our custom split function.
	scan := bufio.NewScanner(sess.port)
	scan.Split(xbeeFrameSplit)

	for scan.Scan() {
		// TODO: check for errors?
		// data is a frame payload - not a full frame.
		data, err := parseFrame(scan.Bytes())
		if err != nil {
			sess.Logger.Warn("error parsing frame", "error", err, "data", data)
			continue
		}
		// data is good, lets parse the frame - using the first byte as the identifier.

		switch XBeeCmd(data[0]) {
		case RxPktType:
			// we parse the data, and push it to the rx buffer.
			//TODO: if we have multiple sources, we need to track them here.
			frame, err := ParseRxFrame(data)
			if err != nil {
				sess.Logger.Warn("error parsing rx packet", "error", err, "data", data)
				break //continue?
			}
			// take the data and write it to our internal rx packet buffer.
			_, err = sess.rxBuf.Write(frame.Payload)
			if err != nil {
				sess.Logger.Warn("error writing data", "error", err, "payload", frame.Payload)
			}

		// the "callback"-style handler. Any received packet with a frame ID should
		// be handled here.
		case TxStatusType, ATCmdResponseType, RemoteCmdRespType: // these take the frame bytes and parse it themselves.
			// we hand it back via the channel. we directly find the ID since it's always
			// the second byte.
			idx := data[1]

			err := sess.ct.ClearMark(idx, data)
			if err != nil {
				// we got a rogue packet lol
				sess.Logger.Warn("rogue frame ID", "id", idx, "error", err)
			}

		default:
			// we don't know what to do with it.
			sess.Logger.Info("unhandled packet type", "type", data[0], "id", data[1])

		}

	}
	// if we get here, the serial port has closed. this is fine.
}

// This implements io.Reader for the UART Session.
func (sess *SerialSession) Read(p []byte) (int, error) {
	// Since we have an rx buffer, we just read from that and return the results.
	return sess.rxBuf.Read(p)
}

func (sess *SerialSession) Write(p []byte) (n int, err error) {
	// we construct a packet - using the conntrack to ensure that the packet is okay.
	// we block - this is more correct.
	idx, ch, err := sess.ct.GetMark()
	if err != nil {
		return
	}
	wf := &TxFrame{
		Id:          idx,
		Destination: BroadcastAddr,
		Payload:     p,
	}
	// write the actual packet

	sess.writeLock.Lock()
	n, err = writeXBeeFrame(sess.port, wf.Bytes())
	sess.writeLock.Unlock()
	if err != nil {
		return
	}

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
func (sess *SerialSession) ATCommand(cmd [2]rune, data []byte, queued bool) ([]byte, error) {
	// we must encode the command, and then create the actual packet.
	// then we send the packet, and wait for the response
	// TODO: how to handle multiple-response-packet AT commands?
	// (mainly Node Discovery ND)

	// get a mark for the frame

	isQuery := len(data) > 0
	idx, ch, err := sess.ct.GetMark()
	if err != nil {
		return nil, err
	}
	rawData := encodeATCommand(cmd, data, idx, queued)

	sess.writeLock.Lock()
	_, err = writeXBeeFrame(sess.port, rawData)
	sess.writeLock.Unlock()

	if err != nil {
		return nil, fmt.Errorf("error writing xbee frame: %w", err)
	}

	// we use the AT command that was provided to decode the frame.
	// Parse stores the response result locally.
	// we parse the base frame ourselves, and if it's okay we pass it
	// to the provided ATCommand

	// TODO: add timeout.
	resp, err := ParseATCmdResponse(<-ch)
	if err != nil {
		return nil, err
	}

	if resp.Status != 0 {
		// sinec ATCmdStatus is a stringer thanks to the generator
		return nil, fmt.Errorf("AT command failed: %v", resp.Status)
	}

	// finally, we use the provided ATCmd interface to unpack the data.
	// this overwrites the values provided, but this should only happen
	// during a query, so this is fine.
	// TODO: skip if not a query command?

	if isQuery {
		return resp.Data, nil
	}

	// it's not a query, and there was no error, so we just plain return
	return nil, nil

}

// Does this need to exist?
func (sess *SerialSession) GetStatus() {
	panic("TODO: implement")
}

// Implement the io.Closer.
func (sess *SerialSession) Close() error {
	return sess.port.Close()
}

func (sess *SerialSession) DiscoverNodes() {
	panic("TODO: implement")
}

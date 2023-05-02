package gotelem

import (
	"errors"
	"math/rand"
	"net"
	"sync"

	"github.com/tinylib/msgp/msgp"
)

// the target architecture is a subscribe function that
// takes a can FILTER. Then the server will emit notifications.
// that contain new can packets as they come in.

// this means that the client should be able to handle
// notify packets on top of response packets.

// we should register handlers. They should handle serialization
// and deserialization on their own. This way we avoid reflect.
// since reflected code can be more complex under the hood.

// ServiceFunc is a RPC service handler.
type ServiceFunc func(params msgp.Raw) (res msgp.Raw, err error)

// RPCConn is a single RPC communication pair.
type RPCConn struct {
	// TODO: use io.readwritecloser?
	conn     net.Conn
	handlers map[string]ServiceFunc

	// indicates what messages we've used.
	// TODO: use a channel to return a response?
	// TODO: lock with mutex
	ct rpcConnTrack
}

// Call intiates an RPC call to a remote method and returns the
// response, or the error, if any.
// TODO: determine signature
// TODO: this should block?
func (rpc *RPCConn) Call(method string, params msgp.Marshaler) (msgp.Raw, error) {

	// TODO: error handling.
	rawParam, _ := params.MarshalMsg([]byte{})

	id, cb := rpc.ct.Claim()

	req := NewRequest(id, method, rawParam)

	w := msgp.NewWriter(rpc.conn)
	req.EncodeMsg(w)

	// block and wait for response.
	resp := <-cb

	return resp.Result, &resp.Error
}

// Notify initiates a notification to a remote method. It does not
// return any information. There is no response from the server.
// This method will not block. An error is returned if there is a local
// problem.
func (rpc *RPCConn) Notify(method string, params msgp.Marshaler) {
	// TODO: return an error if there's a local problem?
	rawParam, _ := params.MarshalMsg([]byte{})

	req := NewNotification(method, rawParam)

	w := msgp.NewWriter(rpc.conn)
	req.EncodeMsg(w)

}

// Register a new handler to be called by the remote side. An error
// is returned if the handler name is already in use.
func (rpc *RPCConn) RegisterHandler(name string, fn ServiceFunc) error {
	// TODO: check if name in use.
	// TODO: mutex lock for sync (or use sync.map?
	rpc.handlers[name] = fn

	return nil
}

// Serve runs the server. It will dispatch goroutines to handle each
// method call. This can (and should in most cases) be run in the background to allow for
// sending and receving on the same connection.
func (rpc *RPCConn) Serve() {

	// construct a stream reader.
	msgReader := msgp.NewReader(rpc.conn)

	// read a request/notification from the connection.

	var rawmsg msgp.Raw = make(msgp.Raw, 0, 4)

	for {
		rawmsg.DecodeMsg(msgReader)

		rpcIntf, _ := parseRPC(rawmsg)

		switch rpcObject := rpcIntf.(type) {
		case Request:
			// the object is a request - we must dispatch a goroutine
			// that will call the handler and also send a return value.
			go rpc.dispatch(rpcObject)
		case Notification:
			go rpc.dispatchNotif(rpcObject)
		case Response:
			cbCh, err := rpc.ct.Clear(rpcObject.MsgId)
			if err != nil {
				// TODO: scream
			}
			cbCh <- rpcObject
		}
	}
}

func (rpc *RPCConn) dispatch(req Request) {

	result, err := rpc.handlers[req.Method](req.Params)

	if err != nil {
		// log the error.
	}
	// construct the response frame.
	var rpcE *RPCError = MakeRPCError(err)

	w := msgp.NewWriter(rpc.conn)
	resBuf := make(msgp.Raw, result.Msgsize())

	result.MarshalMsg(resBuf)

	response := NewResponse(req.MsgId, *rpcE, resBuf)

	response.EncodeMsg(w)

}
func (rpc *RPCConn) dispatchNotif(req Notification) {

	_, err := rpc.handlers[req.Method](req.Params)

	if err != nil {
		// log the error.
	}
	// no need for response.
}

// RPCConntrack is a request-response tracker that is used to connect
// the response to the appropriate caller.
type rpcConnTrack struct {
	ct map[uint32]chan Response // TODO: change the values of the map for callbacks.
	mu sync.RWMutex
}

// Get attempts to get a random mark from the mutex.
func (c *rpcConnTrack) Claim() (uint32, chan Response) {
	// TODO: make this threadsafe.
	var val uint32
	for {

		newVal := rand.Uint32()
		// collision is *rare* - so we just try again.
		// I hope to god you don't saturate this tracker.
		c.mu.RLock()
		if _, exist := c.ct[newVal]; !exist {
			val = newVal
			c.mu.RUnlock()
			break
		}
		c.mu.RUnlock()
	}

	// claim it
	// the channel should be buffered. We only expect one value to go through.
	// so the size is fixed.
	ch := make(chan Response, 1)
	c.mu.Lock()
	c.ct[val] = ch
	c.mu.Unlock()

	return val, ch
}

// Clear deletes the connection from the tracker and returns the channel
// associated with it. The caller can use the channel afterwards
// to send the response.
func (c *rpcConnTrack) Clear(val uint32) (chan Response, error) {
	// TODO: get a lock
	c.mu.RLock()
	ch, ok := c.ct[val]
	c.mu.RUnlock()
	if !ok {
		return nil, errors.New("invalid msg id")
	}
	c.mu.Lock()
	delete(c.ct, val)
	c.mu.Unlock()
	return ch, nil
}

// Next, we define some helper generic functions that can be used to make
// implementing a msg wrapper easier.

type msgpackObject interface {
	msgp.Decodable
	msgp.Encodable
	msgp.MarshalSizer
	msgp.Unmarshaler
}

// MakeService is a generic wrapper function. It takes a function with the signature
// of func(T msgpObject)(R msgpObject, error) where T and R can be *concrete* types.
// and returns a new function that handles conversion to/from msgp.Raw.
// the function returned can be used by the RPCConn as a handler function.
func MakeService[T, R msgpackObject](fn func(T) (R, error)) func(msgp.Raw) (msgp.Raw, error) {
	return func(p msgp.Raw) (msgp.Raw, error) {
		// decode the raw data into a new underlying type.
		var params T
		// TODO: handler errors
		_, err := params.UnmarshalMsg(p)

		if err != nil {
			return nil, err
		}

		// now, call the function fn with the given params, and record the value.

		resp, err := fn(params)

		if err != nil {
			return nil, err
		}

		return resp.MarshalMsg([]byte{})

	}
}

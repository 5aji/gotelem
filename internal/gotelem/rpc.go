package gotelem

import (
	"errors"
	"math/rand"
	"net"
	"sync"

	"github.com/tinylib/msgp/msgp"
	"golang.org/x/exp/slog"
)

// the target architecture is a subscribe function that
// takes a can FILTER. Then the server will emit notifications.
// that contain new can packets as they come in.

// this means that the client should be able to handle
// notify packets on top of response packets.

// we should register handlers. They should handle serialization
// and deserialization on their own. This way we avoid reflect.
// since reflected code can be more complex under the hood.
// to make writing services easier, we can use generic functions
// that convert a normal go function to a serviceFunc

// ServiceFunc is a RPC service handler. It can be created manually,
// or by using the generic MakeService function on a
// `func(msgp.Encoder) (msgp.Deocder, error)`
type ServiceFunc func(params msgp.Raw) (res msgp.Raw, err error)

// RPCConn is a single RPC communication pair. It is used by both the
// "server" aka listener, and client. Dynamic registration of service
// handlers is supported.
type RPCConn struct {
	// TODO: use io.readwritecloser?
	conn     net.Conn
	handlers map[string]ServiceFunc

	ct rpcConnTrack

	slog.Logger
}

// Call intiates an RPC call to a remote method and returns the
// response, or the error, if any.
// TODO: determine signature. Should params be msgp.Raw?
func (rpc *RPCConn) Call(method string, params msgp.Raw) (msgp.Raw, error) {

	// TODO: error handling.

	id, cb := rpc.ct.Claim()

	req := NewRequest(id, method, params)

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
func (rpc *RPCConn) Notify(method string, params msgp.Raw) {
	// TODO: return an error if there's a local problem?

	req := NewNotification(method, params)

	w := msgp.NewWriter(rpc.conn)
	req.EncodeMsg(w)

}

// Register a new handler to be called by the remote side. An error
// is returned if the handler name is already in use.
func (rpc *RPCConn) RegisterHandler(name string, fn ServiceFunc) error {
	// TODO: check if name in use.
	// TODO: mutex lock for sync (or use sync.map?
	rpc.handlers[name] = fn

	rpc.Logger.Info("registered a new handler", "name", name, "fn", fn)

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

		rpcIntf, err := parseRPC(rawmsg)

		if err != nil {
			rpc.Logger.Warn("Could not parse RPC message", "err", err)
			continue
		}

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
				rpc.Logger.Warn("could not get rpc callback", "msgid", rpcObject.MsgId, "err", err)
				continue
			}
			cbCh <- rpcObject
		}
	}
}

func (rpc *RPCConn) dispatch(req Request) {

	result, err := rpc.handlers[req.Method](req.Params)

	if err != nil {
		rpc.Logger.Warn("error dispatching rpc function", "method", req.Method, "err", err)
	}
	// construct the response frame.
	var rpcE *RPCError = MakeRPCError(err)

	w := msgp.NewWriter(rpc.conn)

	response := NewResponse(req.MsgId, *rpcE, result)

	response.EncodeMsg(w)

}
func (rpc *RPCConn) dispatchNotif(req Notification) {

	_, err := rpc.handlers[req.Method](req.Params)

	if err != nil {
		// log the error, but don't do anything about it.
		rpc.Logger.Warn("error dispatching rpc function", "method", req.Method, "err", err)
	}
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
// This function can typically have it's paramters inferred.
func MakeService[T, R msgpackObject](fn func(T) (R, error)) ServiceFunc {
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

// should the RPCConn/method name be baked into the function or should they be
// part of the returned function paramters?

// MakeCaller creates a simple wrapper around a parameter of call. The method name
// and RPC connection can be given to the returned function to make a RPC call on that
// function with the given type parameters.
//
// This function is slightly obtuse compared to MakeBoundCaller but is more flexible
// since you can reuse the same function across multiple connections and method names.
//
// This generic function must always have it's type paratmers declared explicitly.
// They cannot be inferred from the given parameters.
func MakeCaller[T, R msgpackObject]() func(string, T, *RPCConn) (R, error) {
	return func(method string, param T, rpc *RPCConn) (R, error) {

		rawParam, _ := param.MarshalMsg([]byte{})
		rawResponse, err := rpc.Call(method, rawParam)
		if err != nil {
			var emtpyR R
			return emtpyR, err
		}

		var resp R

		_, err = resp.UnmarshalMsg(rawResponse)

		return resp, err
	}
}

// MakeBoundCaller is like MakeCaller, except the RPC connection and method name are
// fixed and cannot be adjusted later. This function is more elegant but less flexible
// than MakeCaller and should be used when performance is not critical.
//
// This generic function must always have it's type paratmers declared explicitly.
// They cannot be inferred from the given parameters.
func MakeBoundCaller[T, R msgpackObject](rpc *RPCConn, method string) func(T) (R, error) {

	return func(param T) (R, error) {
		// encode parameters
		// invoke rpc.Call
		// await response
		// unpack values.
		rawParam, _ := param.MarshalMsg([]byte{})

		rawResponse, err := rpc.Call(method, rawParam)
		if err != nil {
			var emtpyR R
			return emtpyR, err
		}

		var resp R

		_, err = resp.UnmarshalMsg(rawResponse)

		return resp, err

	}
}

func MakeNotifier[T msgpackObject]() func(string, T, *RPCConn) {
	return func(method string, param T, rpc *RPCConn) {
		rawParam, _ := param.MarshalMsg([]byte{})
		rpc.Notify(method, rawParam)
	}
}

/*
mprpc is a simple bidirectional RPC library using the MessagePack-RPC spec.

It fully implements the spec and additionally provides Go `error“ handling by
converting the error to a standard format for other clients.

mprpc does not have a typical server/client designation - both use "handlers",
which expose methods to be called over the network. A "client" would be an
RPCConn which doesn't expose any services, and a "server" would be an RPCConn
that doesn't make any `Call`s to the other side.

This lack of discrete server and client enables mprpc to implement a basic
"streaming" architecture on top of the MessagePack-RPC spec, which does not
include streaming primitives.  Instead, we can provide simple "service handlers"
as a callback/destination for streaming data.

For example, a "client" could subscribe to events from the "server", by
providing a callback service to point events to. Then, the "server" would
Notify() the callback service with the new event as an argument every time it
occured.  While this may be less optimal than protocol-level streaming, it is
far simpler.

# Generic Helper Functions

The idiomatic way to use mprpc is to use the generic functions that are provided
as helpers. They allow the programmer to easily wrap existing functions in a
closure that automatically encodes and decodes the parameters and results to
their MessagePack representations. See the Make* generic functions for more
information.

	// Assume myParam and myResult are MessagePack-enabled structs.
	// Use `msgp` to generate the required functions for them.

	// this is our plain function - we can call it locally to test.
	func myPlainFunction(p myParam) (r myResult, err error)

	// wrapped is a ServiceFunc that can be passed to rpcConn.RegisterHandler
	var wrapped := MakeService(myPlainFunction)

The generic functions allow for flexiblity and elegant code while still keeping
the underlying implementation reflect-free. For more complex functions (i.e
multiple parameters or return types), a second layer of indirection can be used.

There is also a `MakeCaller` function that can make a stub function that handles
encoding the arguments and decoding the response for a remote procedure.
*/
package mprpc

import (
	"errors"
	"io"

	"github.com/tinylib/msgp/msgp"
	"golang.org/x/exp/slog"
)

// ServiceFunc is a RPC service handler.
// It can be created manually, or by using the generic MakeService function on a
//
//	func(msgp.Encoder) (msgp.Decoder, error)
//
// type.
type ServiceFunc func(params msgp.Raw) (res msgp.Raw, err error)

// RPCConn is a single RPC communication pair.
// It is used by both the
// "server" aka listener, and client.
type RPCConn struct {
	// TODO: use io.readwritecloser?
	rwc      io.ReadWriteCloser
	handlers map[string]ServiceFunc

	ct rpcConnTrack

	logger slog.Logger
}

// creates a new RPC connection on top of an io.ReadWriteCloser. Can be
// pre-seeded with handlers.
func NewRPC(rwc io.ReadWriteCloser, logger *slog.Logger, initialHandlers map[string]ServiceFunc) (rpc *RPCConn, err error) {

	rpc = &RPCConn{
		rwc:      rwc,
		handlers: make(map[string]ServiceFunc),
		ct:       NewRPCConnTrack(),
	}
	if initialHandlers != nil {
		for k, v := range initialHandlers {
			rpc.handlers[k] = v
		}
	}

	return

}

// Call intiates an RPC call to a remote method and returns the
// response, or the error, if any. To make calling easier, you can
// construct a "Caller" with MakeCaller
func (rpc *RPCConn) Call(method string, params msgp.Raw) (msgp.Raw, error) {

	// TODO: error handling.

	id, cb := rpc.ct.Claim()

	req := NewRequest(id, method, params)

	w := msgp.NewWriter(rpc.rwc)
	req.EncodeMsg(w)

	// block and wait for response.
	resp := <-cb

	return resp.Result, &resp.Error
}

// Notify initiates a notification to a remote method. It does not
// return any information. There is no response from the server.
// This method will not block nor will it inform the caller if any errors occur.
func (rpc *RPCConn) Notify(method string, params msgp.Raw) {
	// TODO: return an error if there's a local problem?

	req := NewNotification(method, params)

	w := msgp.NewWriter(rpc.rwc)
	req.EncodeMsg(w)

}

// Register a new handler to be called by the remote side. An error
// is returned if the handler name is already in use.
func (rpc *RPCConn) RegisterHandler(name string, fn ServiceFunc) error {
	// TODO: check if name in use.
	// TODO: mutex lock for sync (or use sync.map?
	rpc.handlers[name] = fn

	rpc.logger.Info("registered a new handler", "name", name, "fn", fn)

	return nil
}

// Removes a handler, if it exists. Never errors. No-op if the name
// is not a registered handler.
func (rpc *RPCConn) RemoveHandler(name string) error {
	delete(rpc.handlers, name)
	return nil
}

// Serve runs the server. It will dispatch goroutines to handle each method
// call. This can (and should in most cases) be run in the background to allow
// for sending and receving on the same connection.
func (rpc *RPCConn) Serve() {

	// construct a stream reader.
	msgReader := msgp.NewReader(rpc.rwc)

	// read a request/notification from the connection.

	var rawmsg msgp.Raw = make(msgp.Raw, 0, 4)

	for {
		err := rawmsg.DecodeMsg(msgReader)
		if err != nil {
			if errors.Is(err, io.EOF) {
				rpc.logger.Info("reached EOF, stopping server")
				return
			}
			rpc.logger.Warn("error decoding message", "err", err)
			continue
		}

		rpcIntf, err := parseRPC(rawmsg)

		if err != nil {
			rpc.logger.Warn("Could not parse RPC message", "err", err)
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
				rpc.logger.Warn("could not get rpc callback", "msgid", rpcObject.MsgId, "err", err)
				continue
			}
			cbCh <- rpcObject
		default:
			panic("invalid rpcObject!")
		}
	}
}

// INTERNAL functions for rpcConn

// dispatch is an internal method used to execute a Request sent by the remote:w
func (rpc *RPCConn) dispatch(req Request) {

	result, err := rpc.handlers[req.Method](req.Params)

	if err != nil {
		rpc.logger.Warn("error dispatching rpc function", "method", req.Method, "err", err)
	}
	// construct the response frame.
	var rpcE *RPCError = MakeRPCError(err)

	w := msgp.NewWriter(rpc.rwc)

	response := NewResponse(req.MsgId, *rpcE, result)

	response.EncodeMsg(w)

}

// dispatchNotif is like dispatch, but for Notifications. This means that it never replies,
// even if there is an error.
func (rpc *RPCConn) dispatchNotif(req Notification) {

	_, err := rpc.handlers[req.Method](req.Params)

	if err != nil {
		// log the error, but don't do anything about it.
		rpc.logger.Warn("error dispatching rpc function", "method", req.Method, "err", err)
	}
}

// Next, we define some helper generic functions that can be used to make
// implementing a msg wrapper easier.

// msgpackObject is anything that has implemented all the msgpack interfaces.
type msgpackObject interface {
	msgp.Decodable
	msgp.Encodable
	msgp.MarshalSizer
	msgp.Unmarshaler
}

// MakeService is a generic wrapper function. It takes a function with the signature
// of func(T msgpObject)(R msgpObject, error) where T and R can be *concrete* types.
// and returns a new function that handles conversion to/from msgp.Raw.
// The function returned can be used by the RPCConn as a handler function.
// This function can typically have it's paramters inferred.
func MakeService[T, R msgpackObject](fn func(T) (R, error)) ServiceFunc {
	return func(p msgp.Raw) (msgp.Raw, error) {
		// decode the raw data into a new underlying type.
		var params T

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

		rawParam, err := param.MarshalMsg([]byte{})
		if err != nil {
			var emtpyR R
			return emtpyR, err
		}
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

// MakeNotifier creates a new notification function that notifies the remote
func MakeNotifier[T msgpackObject](method string) func(T, *RPCConn) error {
	return func(param T, rpc *RPCConn) error {
		rawParam, err := param.MarshalMsg([]byte{})
		rpc.Notify(method, rawParam)
		return err
	}
}

package gotelem

import (
	"net"

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
type ServiceFunc func(params msgp.Raw) (res msgp.MarshalSizer, err error)


// RPCConn is a single RPC communication pair.
type RPCConn struct {
	// TODO: use io.readwritecloser?
	conn     net.Conn
	handlers map[string]ServiceFunc

	// indicates what messages we've used.
	// TODO: use a channel to return a response?
	// TODO: lock with mutex
	ct map[uint32]struct{}
}

// Call intiates an RPC call to a remote method and returns the
// response, or the error, if any.
// TODO: determine signature
// TODO: this should block?
func (rpc *RPCConn) Call(method string, params msgp.Marshaler) {

}

// Notify initiates a notification to a remote method. It does not
// return any information. There is no response from the server.
// This method will not block. An error is returned if there is a local
// problem.
func (rpc *RPCConn) Notify(method string, params msgp.Marshaler) {
	// TODO: return an error if there's a local problem?

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

	rawmsg.DecodeMsg(msgReader)

	rpcIntf, err := parseRPC(rawmsg)

	switch rpcObject := rpcIntf.(type) {
	case Request:
		// the object is a request - we must dispatch a goroutine
		// that will call the handler and also send a return value.
		go rpc.dispatch(rpcObject)
	case Notification:
		go rpc.dispatchNotif(rpcObject)
	case Response:
		// TODO: return response to caller.
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

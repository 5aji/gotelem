package gotelem

import (
	"errors"

	"github.com/tinylib/msgp/msgp"
)

// this file is a simple implementation of the msgpack-rpc data formato.
// it also contains an RPC server and client.
// We can port this to python rather easily too.

type RPCType int

const (
	RequestType      RPCType = 0
	ResponseType     RPCType = 1
	NotificationType RPCType = 2
)

// the messagepack RPC spec requires that the RPC wire formts are ordered arrays,
// aka tuples. we can use msgp options to make them tuple automatically,
// based on the order they are declared. This makes the order of these
// structs *critical*! Do not touch!

//go:generate msgp
//msgp:tuple Request
//msgp:tuple Response
//msgp:tuple Notification

// A request is a function call that expects a Response.
type Request struct {
	// should always be zero.
	msgtype RPCType  `msg:"type"`
	MsgId   uint32   `msg:"msgid"`
	Method  string   `msg:"method"`
	Params  msgp.Raw `msg:"params,allownil"`
}

func NewRequest(msgid uint32, method string, params msgp.Raw) *Request {
	return &Request{
		msgtype: 0,
		MsgId:   msgid,
		Method:  method,
		Params:  params,
	}
}

// A response is the result of a function call, or an error.
type Response struct {
	// should always be one.
	msgtype RPCType  `msg:"type"`
	MsgId   uint32   `msg:"msgid"`
	Error   RPCError `msg:"error,allownil"`
	Result  msgp.Raw `msg:"result,allownil"`
}

func NewResponse(msgid uint32, respErr RPCError, res msgp.Raw) *Response {
	return &Response{
		msgtype: 1,
		MsgId:   msgid,
		Error:   respErr,
		Result:  res,
	}
}

// A notification is a function call that does not care if the call
// succeeds and ignores responses.
type Notification struct {
	// should always be *2*
	msgtype RPCType  `msg:"type"`
	Method  string   `msg:"method"`
	Params  msgp.Raw `msg:"params,allownil"`
}

func NewNotification(method string, params msgp.Raw) *Notification {
	return &Notification{
		msgtype: 2,
		Method:  method,
		Params:  params,
	}
}

// todo: should these be functions instead, since they're arrays? and we need to determine the type beforehand.

func getMsgType(b []byte) RPCType {
	size, next, err := msgp.ReadArrayHeaderBytes(b)
	if err != nil {
		panic(err)
	}
	if size == 3 { // hot path for notifications.
		return NotificationType
	}

	vtype, _, err := msgp.ReadIntBytes(next)

	if err != nil {
		panic(err)
	}

	// todo: use readIntf instead? returns a []interface{} and we can map it ourselves...
	return RPCType(vtype)
}

// parseRPC takes a raw message and decodes it based on the first value
// of the array (the type). It returns the decoded object. Callers
// can use a type-switch to determine the type of the data.
func parseRPC(raw msgp.Raw) (interface{}, error) {
	t := getMsgType(raw)

	switch RPCType(t) {

	case RequestType:
		// create and return a request struct.
		req := &Request{}
		_, err := req.UnmarshalMsg(raw)
		return req, err
	case ResponseType:
		res := &Response{}
		_, err := res.UnmarshalMsg(raw)
		return res, err
	case NotificationType:
		notif := &Notification{}
		_, err := notif.UnmarshalMsg(raw)
		return notif, err
	default:
		// uh oh.
		return nil, errors.New("unmatched RPC type")
	}
}

// RPCError is a common RPC error format. It is basically a clone of the
// JSON-RPC error format. We use it so we know what to expect there.

//msgp:tuple RPCError
type RPCError struct {
	Code int
	Desc string
}

// Converts a go error into a RPC error.
func MakeRPCError(err error) *RPCError {
	if err == nil {
		return nil
	}
	return &RPCError{
		Code: -1,
		Desc: err.Error(),
	}
}

func (r *RPCError) Error() string {
	return r.Desc
}

package mprpc

import (
	"errors"

	"github.com/tinylib/msgp/msgp"
)

// this file is a simple implementation of the msgpack-rpc data formats.

// RPCType is the message type that is being sent.
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

// Request represents a function call that expects a Response.
type Request struct {
	// should always be zero.
	msgtype RPCType `msg:"type"`
	// MsgId is used to match a Response with a Request
	MsgId uint32 `msg:"msgid"`
	// Method is the name of the method/service to execute on the remote
	Method string `msg:"method"`
	// Params is the arguments of the method/service. It can be any
	// MessagePack-serializable type.
	Params msgp.Raw `msg:"params,allownil"`
}

func NewRequest(msgid uint32, method string, params msgp.Raw) *Request {
	return &Request{
		msgtype: 0,
		MsgId:   msgid,
		Method:  method,
		Params:  params,
	}
}

// A Response is the result and error given from calling a service.
type Response struct {
	// should always be one.
	msgtype RPCType `msg:"type"`
	// MsgId is an identifier used to match this Response with the Request that created it.
	MsgId uint32 `msg:"msgid"`
	// Error is the error encountered while attempting to execute the method, if any.
	Error RPCError `msg:"error,allownil"`
	// Result is the raw object that was returned by the calling method. It
	// can be any MessagePack-serializable object.
	Result msgp.Raw `msg:"result,allownil"`
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

// getMsgType uses raw messagpack RPC to return the underlying message type from
// the raw array given by b.
func getMsgType(b msgp.Raw) RPCType {
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

//msgp:tuple RPCError

// RPCError is a common RPC error format. It is basically a clone of the
// JSON-RPC error format. We use it so we know what to expect there.
type RPCError struct {
	Code int
	Desc string
}

// Converts a Go error into a RPC error.
func MakeRPCError(err error) *RPCError {
	if err == nil {
		return nil
	}
	return &RPCError{
		Code: -1,
		Desc: err.Error(),
	}
}

// Implements the Error interface for RPCError
func (r *RPCError) Error() string {
	return r.Desc
}

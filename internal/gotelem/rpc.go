package gotelem

import "github.com/tinylib/msgp/msgp"

// this file is a simple implementation of the msgpack-rpc data format.

type RPCType int

const (
	RequestType      RPCType = 0
	ResponseType     RPCType = 1
	NotificationType RPCType = 2
)

//go:generate msgp
//msgp:tuple Request
//msgp:tuple Response
//msgp:tuple Notification

// A request is a function call that expects a Response.
type Request struct {
	// should always be zero.
	msgtype int         `msg:"type"`
	MsgId   uint32      `msg:"msgid"`
	Method  string      `msg:"method"`
	Params  interface{} `msg:"params,allownil"`
}

func NewRequest(msgid uint32, method string, params interface{}) *Request {
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
	msgtype int         `msg:"type"`
	MsgId   uint32      `msg:"msgid"`
	Error   interface{} `msg:"error,allownil"`
	Result  interface{} `msg:"result,allownil"`
}

// A notification is a function call that does not care if the call
// succeeds and ignores responses.
type Notification struct {
	// should always be *2*
	msgtype int         `msg:"type"`
	Method  string      `msg:"method"`
	Params  interface{} `msg:"params,allownil"`
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

func parseRPC(raw msgp.Raw) interface{} {
	t := getMsgType(raw)

	if t == RequestType {

	}
	return nil
}

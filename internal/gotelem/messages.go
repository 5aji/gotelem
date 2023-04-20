package gotelem

import (
	"github.com/tinylib/msgp/msgp"
)

// a body is a thing that can get a type, which we put in the header.
// we use the header to store metadata too
type Body interface {
	GetType() string
	msgp.Marshaler
}

//go:generate msgp
type Data struct {
	Header map[string]string `msg:"header"`
	Body   msgp.Raw          `msg:"body"`
}

type CanBody struct {
	Id      uint32 `msg:"id"`
	Payload []byte `msg:"data"`
	Source  string `msg:"src"`
}

func (*CanBody) GetType() string {
	return "canp"
}

// A status contains information about the running application.
// mainly internal battery percentage.
type StatusBody struct {
	BatteryPct float32 `msg:"batt"`
	ErrCode    int16   `msg:"err"` // 0 is good.
}

func (*StatusBody) GetType() string {
	return "status"
}

// takes anything that has a GetType() string method and packs it up.
func NewData(body Body) (*Data, error) {
	data := &Data{}

	data.Header["type"] = body.GetType()

	// add other metadata here.
	data.Header["ver"] = "0.0.1"

	data.Header["test"] = "mesg"

	rawBody, err := body.MarshalMsg(nil)
	if err != nil {
		return nil, err
	}

	data.Body = rawBody

	return data, nil
}

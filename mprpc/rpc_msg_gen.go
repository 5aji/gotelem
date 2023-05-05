package mprpc

// Code generated by github.com/tinylib/msgp DO NOT EDIT.

import (
	"github.com/tinylib/msgp/msgp"
)

// DecodeMsg implements msgp.Decodable
func (z *Notification) DecodeMsg(dc *msgp.Reader) (err error) {
	var zb0001 uint32
	zb0001, err = dc.ReadArrayHeader()
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	if zb0001 != 2 {
		err = msgp.ArrayError{Wanted: 2, Got: zb0001}
		return
	}
	z.Method, err = dc.ReadString()
	if err != nil {
		err = msgp.WrapError(err, "Method")
		return
	}
	err = z.Params.DecodeMsg(dc)
	if err != nil {
		err = msgp.WrapError(err, "Params")
		return
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *Notification) EncodeMsg(en *msgp.Writer) (err error) {
	// array header, size 2
	err = en.Append(0x92)
	if err != nil {
		return
	}
	err = en.WriteString(z.Method)
	if err != nil {
		err = msgp.WrapError(err, "Method")
		return
	}
	err = z.Params.EncodeMsg(en)
	if err != nil {
		err = msgp.WrapError(err, "Params")
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *Notification) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// array header, size 2
	o = append(o, 0x92)
	o = msgp.AppendString(o, z.Method)
	o, err = z.Params.MarshalMsg(o)
	if err != nil {
		err = msgp.WrapError(err, "Params")
		return
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Notification) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadArrayHeaderBytes(bts)
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	if zb0001 != 2 {
		err = msgp.ArrayError{Wanted: 2, Got: zb0001}
		return
	}
	z.Method, bts, err = msgp.ReadStringBytes(bts)
	if err != nil {
		err = msgp.WrapError(err, "Method")
		return
	}
	bts, err = z.Params.UnmarshalMsg(bts)
	if err != nil {
		err = msgp.WrapError(err, "Params")
		return
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *Notification) Msgsize() (s int) {
	s = 1 + msgp.StringPrefixSize + len(z.Method) + z.Params.Msgsize()
	return
}

// DecodeMsg implements msgp.Decodable
func (z *RPCError) DecodeMsg(dc *msgp.Reader) (err error) {
	var zb0001 uint32
	zb0001, err = dc.ReadArrayHeader()
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	if zb0001 != 2 {
		err = msgp.ArrayError{Wanted: 2, Got: zb0001}
		return
	}
	z.Code, err = dc.ReadInt()
	if err != nil {
		err = msgp.WrapError(err, "Code")
		return
	}
	z.Desc, err = dc.ReadString()
	if err != nil {
		err = msgp.WrapError(err, "Desc")
		return
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z RPCError) EncodeMsg(en *msgp.Writer) (err error) {
	// array header, size 2
	err = en.Append(0x92)
	if err != nil {
		return
	}
	err = en.WriteInt(z.Code)
	if err != nil {
		err = msgp.WrapError(err, "Code")
		return
	}
	err = en.WriteString(z.Desc)
	if err != nil {
		err = msgp.WrapError(err, "Desc")
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z RPCError) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// array header, size 2
	o = append(o, 0x92)
	o = msgp.AppendInt(o, z.Code)
	o = msgp.AppendString(o, z.Desc)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *RPCError) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadArrayHeaderBytes(bts)
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	if zb0001 != 2 {
		err = msgp.ArrayError{Wanted: 2, Got: zb0001}
		return
	}
	z.Code, bts, err = msgp.ReadIntBytes(bts)
	if err != nil {
		err = msgp.WrapError(err, "Code")
		return
	}
	z.Desc, bts, err = msgp.ReadStringBytes(bts)
	if err != nil {
		err = msgp.WrapError(err, "Desc")
		return
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z RPCError) Msgsize() (s int) {
	s = 1 + msgp.IntSize + msgp.StringPrefixSize + len(z.Desc)
	return
}

// DecodeMsg implements msgp.Decodable
func (z *RPCType) DecodeMsg(dc *msgp.Reader) (err error) {
	{
		var zb0001 int
		zb0001, err = dc.ReadInt()
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		(*z) = RPCType(zb0001)
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z RPCType) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteInt(int(z))
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z RPCType) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	o = msgp.AppendInt(o, int(z))
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *RPCType) UnmarshalMsg(bts []byte) (o []byte, err error) {
	{
		var zb0001 int
		zb0001, bts, err = msgp.ReadIntBytes(bts)
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		(*z) = RPCType(zb0001)
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z RPCType) Msgsize() (s int) {
	s = msgp.IntSize
	return
}

// DecodeMsg implements msgp.Decodable
func (z *Request) DecodeMsg(dc *msgp.Reader) (err error) {
	var zb0001 uint32
	zb0001, err = dc.ReadArrayHeader()
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	if zb0001 != 3 {
		err = msgp.ArrayError{Wanted: 3, Got: zb0001}
		return
	}
	z.MsgId, err = dc.ReadUint32()
	if err != nil {
		err = msgp.WrapError(err, "MsgId")
		return
	}
	z.Method, err = dc.ReadString()
	if err != nil {
		err = msgp.WrapError(err, "Method")
		return
	}
	err = z.Params.DecodeMsg(dc)
	if err != nil {
		err = msgp.WrapError(err, "Params")
		return
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *Request) EncodeMsg(en *msgp.Writer) (err error) {
	// array header, size 3
	err = en.Append(0x93)
	if err != nil {
		return
	}
	err = en.WriteUint32(z.MsgId)
	if err != nil {
		err = msgp.WrapError(err, "MsgId")
		return
	}
	err = en.WriteString(z.Method)
	if err != nil {
		err = msgp.WrapError(err, "Method")
		return
	}
	err = z.Params.EncodeMsg(en)
	if err != nil {
		err = msgp.WrapError(err, "Params")
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *Request) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// array header, size 3
	o = append(o, 0x93)
	o = msgp.AppendUint32(o, z.MsgId)
	o = msgp.AppendString(o, z.Method)
	o, err = z.Params.MarshalMsg(o)
	if err != nil {
		err = msgp.WrapError(err, "Params")
		return
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Request) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadArrayHeaderBytes(bts)
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	if zb0001 != 3 {
		err = msgp.ArrayError{Wanted: 3, Got: zb0001}
		return
	}
	z.MsgId, bts, err = msgp.ReadUint32Bytes(bts)
	if err != nil {
		err = msgp.WrapError(err, "MsgId")
		return
	}
	z.Method, bts, err = msgp.ReadStringBytes(bts)
	if err != nil {
		err = msgp.WrapError(err, "Method")
		return
	}
	bts, err = z.Params.UnmarshalMsg(bts)
	if err != nil {
		err = msgp.WrapError(err, "Params")
		return
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *Request) Msgsize() (s int) {
	s = 1 + msgp.Uint32Size + msgp.StringPrefixSize + len(z.Method) + z.Params.Msgsize()
	return
}

// DecodeMsg implements msgp.Decodable
func (z *Response) DecodeMsg(dc *msgp.Reader) (err error) {
	var zb0001 uint32
	zb0001, err = dc.ReadArrayHeader()
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	if zb0001 != 3 {
		err = msgp.ArrayError{Wanted: 3, Got: zb0001}
		return
	}
	z.MsgId, err = dc.ReadUint32()
	if err != nil {
		err = msgp.WrapError(err, "MsgId")
		return
	}
	var zb0002 uint32
	zb0002, err = dc.ReadArrayHeader()
	if err != nil {
		err = msgp.WrapError(err, "Error")
		return
	}
	if zb0002 != 2 {
		err = msgp.ArrayError{Wanted: 2, Got: zb0002}
		return
	}
	z.Error.Code, err = dc.ReadInt()
	if err != nil {
		err = msgp.WrapError(err, "Error", "Code")
		return
	}
	z.Error.Desc, err = dc.ReadString()
	if err != nil {
		err = msgp.WrapError(err, "Error", "Desc")
		return
	}
	err = z.Result.DecodeMsg(dc)
	if err != nil {
		err = msgp.WrapError(err, "Result")
		return
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *Response) EncodeMsg(en *msgp.Writer) (err error) {
	// array header, size 3
	err = en.Append(0x93)
	if err != nil {
		return
	}
	err = en.WriteUint32(z.MsgId)
	if err != nil {
		err = msgp.WrapError(err, "MsgId")
		return
	}
	// array header, size 2
	err = en.Append(0x92)
	if err != nil {
		return
	}
	err = en.WriteInt(z.Error.Code)
	if err != nil {
		err = msgp.WrapError(err, "Error", "Code")
		return
	}
	err = en.WriteString(z.Error.Desc)
	if err != nil {
		err = msgp.WrapError(err, "Error", "Desc")
		return
	}
	err = z.Result.EncodeMsg(en)
	if err != nil {
		err = msgp.WrapError(err, "Result")
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *Response) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// array header, size 3
	o = append(o, 0x93)
	o = msgp.AppendUint32(o, z.MsgId)
	// array header, size 2
	o = append(o, 0x92)
	o = msgp.AppendInt(o, z.Error.Code)
	o = msgp.AppendString(o, z.Error.Desc)
	o, err = z.Result.MarshalMsg(o)
	if err != nil {
		err = msgp.WrapError(err, "Result")
		return
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Response) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadArrayHeaderBytes(bts)
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	if zb0001 != 3 {
		err = msgp.ArrayError{Wanted: 3, Got: zb0001}
		return
	}
	z.MsgId, bts, err = msgp.ReadUint32Bytes(bts)
	if err != nil {
		err = msgp.WrapError(err, "MsgId")
		return
	}
	var zb0002 uint32
	zb0002, bts, err = msgp.ReadArrayHeaderBytes(bts)
	if err != nil {
		err = msgp.WrapError(err, "Error")
		return
	}
	if zb0002 != 2 {
		err = msgp.ArrayError{Wanted: 2, Got: zb0002}
		return
	}
	z.Error.Code, bts, err = msgp.ReadIntBytes(bts)
	if err != nil {
		err = msgp.WrapError(err, "Error", "Code")
		return
	}
	z.Error.Desc, bts, err = msgp.ReadStringBytes(bts)
	if err != nil {
		err = msgp.WrapError(err, "Error", "Desc")
		return
	}
	bts, err = z.Result.UnmarshalMsg(bts)
	if err != nil {
		err = msgp.WrapError(err, "Result")
		return
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *Response) Msgsize() (s int) {
	s = 1 + msgp.Uint32Size + 1 + msgp.IntSize + msgp.StringPrefixSize + len(z.Error.Desc) + z.Result.Msgsize()
	return
}
package codec

import (
	"io"
)

/*
	RPC call

err := client.Call("Arith.Multiply", args, &reply)

服务名 + 方法名 + 参数
*/
type Header struct {
	ServiceMethod string // format "Service.Method"

	Seq   uint64 // sequence number chose by client
	Error string
}

type Codec interface {
	io.Closer

	ReadHeader(*Header) error
	ReadBody(any) error
	Write(*Header, any) error
}

type NewCodecFunc func(io.ReadWriteCloser) Codec

type Type string

const (
	GobType  Type = "applocation/gob"
	JsonType Type = "application/json" // not implemented
)

var NewCodecFuncMap map[Type]NewCodecFunc

func init() {
	NewCodecFuncMap = make(map[Type]NewCodecFunc)
	NewCodecFuncMap[GobType] = NewGobCodec
}

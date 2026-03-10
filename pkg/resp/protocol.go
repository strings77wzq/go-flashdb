package resp

import (
	"bytes"
	"strconv"
)

// RESP协议类型标记
const (
	SimpleStringPrefix = '+'
	ErrorPrefix        = '-'
	IntegerPrefix      = ':'
	BulkPrefix         = '$'
	ArrayPrefix        = '*'
	CRLF               = "\r\n"
)

// Reply 通用回复接口
type Reply interface {
	ToBytes() []byte
}

// SimpleStringReply 简单字符串回复
type SimpleStringReply struct {
	Str string
}

func (r *SimpleStringReply) ToBytes() []byte {
	return []byte(string(SimpleStringPrefix) + r.Str + CRLF)
}

// ErrorReply 错误回复
type ErrorReply struct {
	Err string
}

func (r *ErrorReply) ToBytes() []byte {
	return []byte(string(ErrorPrefix) + r.Err + CRLF)
}

// IntegerReply 整数回复
type IntegerReply struct {
	Num int64
}

func (r *IntegerReply) ToBytes() []byte {
	return []byte(string(IntegerPrefix) + strconv.FormatInt(r.Num, 10) + CRLF)
}

// BulkReply 批量回复
type BulkReply struct {
	Arg []byte
}

func (r *BulkReply) ToBytes() []byte {
	if r.Arg == nil {
		return []byte("$-1\r\n")
	}
	return []byte(string(BulkPrefix) + strconv.Itoa(len(r.Arg)) + CRLF + string(r.Arg) + CRLF)
}

// ArrayReply 数组回复
type ArrayReply struct {
	Replies []Reply
}

func (r *ArrayReply) ToBytes() []byte {
	buf := bytes.NewBuffer(make([]byte, 0, 64))
	buf.WriteByte(ArrayPrefix)
	buf.WriteString(strconv.Itoa(len(r.Replies)))
	buf.WriteString(CRLF)
	for _, reply := range r.Replies {
		buf.Write(reply.ToBytes())
	}
	return buf.Bytes()
}

// 常用快捷回复
var (
	OkReply        = &SimpleStringReply{Str: "OK"}
	NilBulkReply   = &BulkReply{Arg: nil}
	NullArrayReply = &ArrayReply{Replies: nil}
	PongReply      = &SimpleStringReply{Str: "PONG"}
)

// NewErrorReply 创建错误回复
func NewErrorReply(err string) *ErrorReply {
	return &ErrorReply{Err: err}
}

// NewBulkReply 创建批量回复
func NewBulkReply(arg []byte) *BulkReply {
	return &BulkReply{Arg: arg}
}

// NewArrayReply 创建数组回复
func NewArrayReply(replies []Reply) *ArrayReply {
	return &ArrayReply{Replies: replies}
}

package reply

import (
	"bytes"
	"go-redis/interface/resp"
	"strconv"
)

var (
	nullBulkReplyBytes = []byte("$-1")
	CRLF               = "\r\n"
)

type BulkReply struct {
	Arg []byte
}

type MultiBulkReply struct {
	Args [][]byte
}

type StatusReply struct {
	Status string
}

type IntReply struct {
	Code int64
}

func MakeBulkReply(arg []byte) *BulkReply {
	return &BulkReply{Arg: arg}
}

func MakeMultiBulkReply(arg [][]byte) *MultiBulkReply {
	return &MultiBulkReply{Args: arg}
}

func MakeStatusReply(status string) *StatusReply {
	return &StatusReply{Status: status}
}

func MakeIntReply(code int64) *IntReply {
	return &IntReply{Code: code}
}

func (r *BulkReply) ToBytes() []byte {
	if len(r.Arg) == 0 {
		return nullBulkReplyBytes
	}
	return []byte("$" + strconv.Itoa(len(r.Arg)) + CRLF + string(r.Arg) + CRLF)
}

func (r *MultiBulkReply) ToBytes() []byte {
	argLen := len(r.Args)

	var buf bytes.Buffer
	buf.WriteString("*" + strconv.Itoa(argLen) + CRLF)
	for _, arg := range r.Args {
		if arg == nil {
			buf.WriteString(string(nullBulkReplyBytes) + CRLF)
		} else {
			buf.WriteString("$" + strconv.Itoa(len(arg)) + CRLF + string(arg) + CRLF)
		}
	}
	return buf.Bytes()
}

func (r *StatusReply) ToBytes() []byte {
	return []byte("+" + r.Status + CRLF)
}

func (r *IntReply) ToBytes() []byte {
	return []byte(":" + strconv.FormatInt(r.Code, 10) + CRLF)
}

type ErrorReply interface {
	Error() string
	ToBytes() []byte
}

type StandardErrReply struct {
	Status string
}

func MakeErrReply(status string) *StandardErrReply {
	return &StandardErrReply{Status: status}
}

func (r *StandardErrReply) ToBytes() []byte {
	return []byte("-" + r.Status + CRLF)
}

func (r *StandardErrReply) Error() string {
	return r.Status
}

func IsErrReply(reply resp.Reply) bool {
	return reply.ToBytes()[0] == '-'
}

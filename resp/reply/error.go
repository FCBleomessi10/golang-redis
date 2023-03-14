package reply

var (
	unknownErrBytes      = []byte("-Err unknown\r\n")
	syntaxErrBytes       = []byte("-Err syntax error\r\n")
	wrongTypeErrReply    = []byte("-WRONGTYPE Operation against a key holding the wrong kind of value\r\n")
	theSyntaxErrReply    = &SyntaxErrReply{}
	theWrongTypeErrReply = &WrongTypeErrReply{}
)

type UnknownErrReply struct {
}

type ArgNumErrReply struct {
	Cmd string
}

type SyntaxErrReply struct {
}

type WrongTypeErrReply struct {
}

type ProtocolErrReply struct {
	Msg string
}

func MakeArgNumErrReply(cmd string) *ArgNumErrReply {
	return &ArgNumErrReply{Cmd: cmd}
}

func MakeSyntaxErrReply() *SyntaxErrReply {
	return theSyntaxErrReply
}

func MakeWrongTypeErrReply() *WrongTypeErrReply {
	return theWrongTypeErrReply
}

func (u *UnknownErrReply) Error() string {
	return "Err unknown"
}

func (u *UnknownErrReply) ToBytes() []byte {
	return unknownErrBytes
}

func (r *ArgNumErrReply) Error() string {
	return "-ERR wrong number of arguments for '" + r.Cmd + "' command\r\n"
}

func (r *ArgNumErrReply) ToBytes() []byte {
	return []byte("-ERR wrong number of arguments for '" + r.Cmd + "' command\r\n")
}

func (r *SyntaxErrReply) Error() string {
	return "Err syntax error"
}

func (r *SyntaxErrReply) ToBytes() []byte {
	return syntaxErrBytes
}

func (r *WrongTypeErrReply) Error() string {
	return "-WRONGTYPE Operation against a key holding the wrong kind of value\r\n"
}

func (r *WrongTypeErrReply) ToBytes() []byte {
	return wrongTypeErrReply
}

func (r *ProtocolErrReply) Error() string {
	return "-ERR Protocol error: '" + r.Msg + "'\r\n"
}

func (r *ProtocolErrReply) ToBytes() []byte {
	return []byte("-ERR Protocol error: '" + r.Msg + "'\r\n'")
}

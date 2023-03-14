package reply

var (
	pongBytes              = []byte("+PONG\r\n")
	okBytes                = []byte("+OK\r\n")
	nullBulkBytes          = []byte("$-1\r\n")
	emptyMultiBulkBytes    = []byte("*0\r\n")
	noBytes                = []byte("")
	thePongReply           = new(PongReply) // 持有一个固定的reply, 不用每次创建对象, 节省内存
	theOkReply             = new(OkReply)
	theNullBulkReply       = new(NullBulkReply)
	theEmptyMultiBulkReply = new(EmptyMultiBulkReply)
	theNoReply             = new(NoReply)
)

type PongReply struct {
}

type OkReply struct {
}

type NullBulkReply struct {
}

type EmptyMultiBulkReply struct {
}

type NoReply struct {
}

func MakePongReply() *PongReply {
	return thePongReply
}

func MakeOkReply() *OkReply {
	return theOkReply
}

func MakeNullBulkReply() *NullBulkReply {
	return theNullBulkReply
}

func MakeEmptyMultiBulkReply() *EmptyMultiBulkReply {
	return theEmptyMultiBulkReply
}

func MakeNoReply() *NoReply {
	return theNoReply
}

func (r *PongReply) ToBytes() []byte {
	return pongBytes
}

func (r *OkReply) ToBytes() []byte {
	return okBytes
}

func (r *NullBulkReply) ToBytes() []byte {
	return nullBulkBytes
}

func (r *EmptyMultiBulkReply) ToBytes() []byte {
	return emptyMultiBulkBytes
}

func (r *NoReply) ToBytes() []byte {
	return noBytes
}

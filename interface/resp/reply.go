package resp

type Reply interface {
	ToBytes() []byte // 把回复的内容转为字节(tcp协议就是来回写字节)
}

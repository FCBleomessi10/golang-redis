package database

import "go-redis/interface/resp"

type CmdLine = [][]byte

type DataBase interface {
	Exec(client resp.Connection, args [][]byte) resp.Reply
	Close()
	AfterClientClose(c resp.Connection)
}

// DataEntity 指代redis所有数据结构
type DataEntity struct {
	Data interface{}
}

package handler

import (
	"context"
	"go-redis/cluster"
	"go-redis/config"
	"go-redis/database"
	databaseface "go-redis/interface/database"
	"go-redis/lib/logger"
	"go-redis/lib/sync/atomic"
	"go-redis/resp/connection"
	"go-redis/resp/parser"
	"go-redis/resp/reply"
	"io"
	"net"
	"strings"
	"sync"
)

var (
	unknownErrReplyBytes = []byte("-ERR unknown\r\n")
)

type RespHandler struct {
	activeConn sync.Map
	db         databaseface.DataBase
	closing    atomic.Boolean
}

func MakeHandler() *RespHandler {
	var db databaseface.DataBase
	if config.Properties.Self != "" && len(config.Properties.Peers) > 0 {
		db = cluster.MakeClusterDatabase()
	} else {
		db = database.NewStandaloneDatabase()
	}
	return &RespHandler{db: db}
}

// 关闭一个client
func (r *RespHandler) closeClient(client *connection.Connection) {
	_ = client.Close()
	r.db.AfterClientClose(client)
	r.activeConn.Delete(client)
}

func (r *RespHandler) Handle(ctx context.Context, conn net.Conn) {
	if r.closing.Get() {
		_ = conn.Close()
	}

	client := connection.NewConn(conn)
	r.activeConn.Store(client, struct{}{})
	ch := parser.ParseStream(conn)
	for payload := range ch {
		// 异常逻辑
		if payload.Err != nil {
			if payload.Err == io.EOF ||
				payload.Err == io.ErrUnexpectedEOF ||
				strings.Contains(payload.Err.Error(), "use of closed network connection") {
				r.closeClient(client)
				logger.Info("connection closed: " + client.RemoteAddr().String())
				return
			}
			errReply := reply.MakeErrReply(payload.Err.Error()) // 协议错误
			err := client.Write(errReply.ToBytes())
			if err != nil {
				r.closeClient(client)
				logger.Info("connection closed: " + client.RemoteAddr().String())
				return
			}
			continue
		}
		// 正常逻辑
		if payload.Data == nil {
			continue
		}
		re, ok := payload.Data.(*reply.MultiBulkReply)
		if !ok {
			logger.Error("require multi bulk reply")
			continue
		}

		result := r.db.Exec(client, re.Args)
		if result != nil {
			_ = client.Write(result.ToBytes())
		} else {
			_ = client.Write(unknownErrReplyBytes)
		}
	}
}

// Close 关闭所有client
func (r *RespHandler) Close() error {
	logger.Info("handler shutting down")
	r.closing.Set(true)
	r.activeConn.Range(func(key, value interface{}) bool {
		client := key.(*connection.Connection)
		_ = client.Close()
		return true
	})
	r.db.Close()
	return nil
}

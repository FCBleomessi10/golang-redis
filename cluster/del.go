package cluster

import (
	"go-redis/interface/resp"
	"go-redis/resp/reply"
)

func Del(cluster *ClusterDatabase, c resp.Connection, cmdArgs [][]byte) resp.Reply {
	replies := cluster.broadcast(c, cmdArgs)
	var errReply reply.ErrorReply
	var deleted int64 = 0

	for _, r := range replies {
		if reply.IsErrReply(r) {
			errReply = r.(reply.ErrorReply)
			break
		}

		intReply, ok := r.(*reply.IntReply)
		if !ok {
			errReply = reply.MakeErrReply("error")
		}
		deleted += intReply.Code
	}

	if errReply == nil {
		return reply.MakeIntReply(deleted)
	}
	return reply.MakeErrReply("error: " + errReply.Error())
}

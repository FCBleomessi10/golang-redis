package cluster

import (
	"go-redis/interface/resp"
	"go-redis/resp/reply"
)

func Rename(cluster *ClusterDatabase, c resp.Connection, cmdArgs [][]byte) resp.Reply {
	if len(cmdArgs) != 3 {
		return reply.MakeErrReply("ERR Wrong number args")
	}

	src, dest := string(cmdArgs[1]), string(cmdArgs[2])
	srcPeer := cluster.peerPicker.PickNode(src)
	destPeer := cluster.peerPicker.PickNode(dest)

	if srcPeer != destPeer {
		return reply.MakeErrReply("ERR rename must within one peer")
	}
	return cluster.relay(srcPeer, c, cmdArgs)
}

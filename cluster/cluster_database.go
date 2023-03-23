package cluster

import (
	"context"
	pool "github.com/jolestar/go-commons-pool/v2"
	"go-redis/config"
	"go-redis/database"
	"go-redis/interface/resp"
	"go-redis/lib/consistenthash"
	"go-redis/lib/logger"
	"go-redis/resp/reply"
	"strings"
)

// 用户请求的转发, 需要连接池. 这样高并发时的请求不需要排队

type ClusterDatabase struct {
	self string

	nodes          []string
	peerPicker     *consistenthash.NodeMap
	peerConnection map[string]*pool.ObjectPool
	db             *database.StandaloneDatabase
}

func MakeClusterDatabase() *ClusterDatabase {
	cluster := &ClusterDatabase{
		self:           config.Properties.Self,
		peerPicker:     consistenthash.NewNodeMap(nil),
		peerConnection: make(map[string]*pool.ObjectPool),
		db:             database.NewStandaloneDatabase(),
	}

	nodes := make([]string, 0, len(config.Properties.Peers)+1)
	for _, peer := range config.Properties.Peers {
		nodes = append(nodes, peer)
	}
	nodes = append(nodes, config.Properties.Self)

	cluster.peerPicker.AddNode(nodes...)

	ctx := context.Background()
	for _, peer := range config.Properties.Peers {
		cluster.peerConnection[peer] = pool.NewObjectPoolWithDefaultConfig(ctx, &connectionFactory{Peer: peer})
	}

	cluster.nodes = nodes
	return cluster
}

type CmdFunc func(*ClusterDatabase, resp.Connection, [][]byte) resp.Reply

var router = makeRouter()

func (cluster *ClusterDatabase) Exec(client resp.Connection, args [][]byte) (result resp.Reply) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error(err)
			result = &reply.UnknownErrReply{}
		}
	}()

	cmdName := strings.ToLower(string(args[0]))
	cmdFunc, ok := router[cmdName]
	if !ok {
		reply.MakeErrReply("not supported cmd")
	}
	result = cmdFunc(cluster, client, args)
	return
}

func (cluster *ClusterDatabase) Close() {
	cluster.db.Close()
}

func (cluster *ClusterDatabase) AfterClientClose(c resp.Connection) {
	cluster.db.AfterClientClose(c)
}

package cluster

import "go-redis/interface/resp"

func makeRouter() map[string]CmdFunc {
	routerMap := make(map[string]CmdFunc)
	routerMap["exists"] = defaultFunc
	routerMap["type"] = defaultFunc
	routerMap["set"] = defaultFunc
	routerMap["setnx"] = defaultFunc
	routerMap["get"] = defaultFunc
	routerMap["getnx"] = defaultFunc
	routerMap["ping"] = ping
	routerMap["rename"] = Rename
	routerMap["renamenx"] = Rename
	routerMap["flushdb"] = flushdb
	routerMap["del"] = Del
	routerMap["select"] = execSelect
	return routerMap
}

func defaultFunc(cluster *ClusterDatabase, c resp.Connection, cmdArgs [][]byte) resp.Reply {
	key := string(cmdArgs[1])
	peer := cluster.peerPicker.PickNode(key)
	return cluster.relay(peer, c, cmdArgs)
}

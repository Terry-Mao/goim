package main

import (
	"strconv"

	log "code.google.com/p/log4go"
	"github.com/Terry-Mao/goim/define"
	"github.com/Terry-Mao/goim/libs/hash/ketama"
	inet "github.com/Terry-Mao/goim/libs/net"
	rproto "github.com/Terry-Mao/goim/proto/router"
	rpc "github.com/Terry-Mao/protorpc"
)

var (
	routerServiceMap = map[string]**rpc.Client{}
	routerRing       *ketama.HashRing
)

const (
	routerService           = "RouterRPC"
	routerServiceConnect    = "RouterRPC.Connect"
	routerServiceDisconnect = "RouterRPC.Disconnect"
	routerServiceMGet       = "RouterRPC.MGet"
	routerServiceGetAll     = "RouterRPC.GetAll"
)

func InitRouter() (err error) {
	var (
		network, addr string
	)
	routerRing = ketama.NewRing(ketama.Base)
	for serverId, addrs := range Conf.RouterRPCAddrs {
		// WARN r must every recycle changed for reconnect
		var (
			r          *rpc.Client
			routerQuit = make(chan struct{}, 1)
		)
		if network, addr, err = inet.ParseNetwork(addrs); err != nil {
			log.Error("inet.ParseNetwork() error(%v)", err)
			return
		}
		r, err = rpc.Dial(network, addr)
		if err != nil {
			log.Error("rpc.Dial(\"%s\", \"%s\") error(%s)", network, addr, err)
		}
		go rpc.Reconnect(&r, routerQuit, network, addr)
		log.Debug("router rpc addr:%s connect", addr)
		routerServiceMap[serverId] = &r
		routerRing.AddNode(serverId, 1)
	}
	routerRing.Bake()
	return
}

func getRouters() map[string]**rpc.Client {
	return routerServiceMap
}

func getRouterByServer(server string) (*rpc.Client, error) {
	if client, ok := routerServiceMap[server]; !ok || *client == nil {
		return nil, define.ErrRouter
	} else {
		return *client, nil
	}
}

func getRouterByUID(userID int64) (*rpc.Client, error) {
	return getRouterByServer(routerRing.Hash(strconv.FormatInt(userID, 10)))
}

func getRouterNode(userID int64) string {
	return routerRing.Hash(strconv.FormatInt(userID, 10))
}

// divide userIds to corresponding
// response: map[nodes]userIds
func divideToRouter(userIds []int64) map[string][]int64 {
	var (
		m    = map[string][]int64{}
		node string
	)
	for i := 0; i < len(userIds); i++ {
		node = getRouterNode(userIds[i])
		ids, ok := m[node]
		if !ok {
			ids = []int64{userIds[i]}
		} else {
			ids = append(ids, userIds[i])
		}
		m[node] = ids
	}
	return m
}

func connect(userID int64, server int32) (seq int32, err error) {
	var client *rpc.Client
	if client, err = getRouterByUID(userID); err != nil {
		return
	}
	arg := &rproto.ConnArg{UserId: userID, Server: server}
	reply := &rproto.ConnReply{}
	if err = client.Call(routerServiceConnect, arg, reply); err != nil {
		log.Error("c.Call(\"%s\",\"%v\") error(%s)", routerServiceConnect, arg, err)
	} else {
		seq = reply.Seq
	}
	return
}

func disconnect(userID int64, seq int32) (has bool, err error) {
	var client *rpc.Client
	if client, err = getRouterByUID(userID); err != nil {
		return
	}
	arg := &rproto.DisconnArg{UserId: userID, Seq: seq}
	reply := &rproto.DisconnReply{}
	if err = client.Call(routerServiceDisconnect, arg, reply); err != nil {
		log.Error("c.Call(\"%s\",\"%v\") error(%s)", routerServiceDisconnect, *arg, err)
	} else {
		has = reply.Has
	}
	return
}

func getSubkeys(serverId string, userIds []int64) (reply *rproto.MGetReply, err error) {
	var client *rpc.Client
	if client, err = getRouterByServer(serverId); err != nil {
		return
	}
	arg := &rproto.MGetArg{UserIds: userIds}
	reply = &rproto.MGetReply{}
	if err = client.Call(routerServiceMGet, arg, reply); err != nil {
		log.Error("client.Call(\"%s\",\"%v\") error(%s)", routerServiceMGet, arg, err)
	}
	return
}

func getAllSubkeys(serverId string) (reply *rproto.GetAllReply, err error) {
	var client *rpc.Client
	if client, err = getRouterByServer(serverId); err != nil {
		return
	}
	reply = &rproto.GetAllReply{}
	if err = client.Call(routerServiceGetAll, nil, reply); err != nil {
		log.Error("client.Call(\"%s\") error(%s)", routerServiceGetAll, err)
	}
	return
}

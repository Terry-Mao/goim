package main

import (
	log "code.google.com/p/log4go"
	"github.com/Terry-Mao/goim/libs/hash/ketama"
	inet "github.com/Terry-Mao/goim/libs/net"
	rproto "github.com/Terry-Mao/goim/proto/router"
	rpc "github.com/Terry-Mao/protorpc"
	"strconv"
)

var (
	routerServiceMap = map[string]*rpc.Client{}
	routerQuit       = make(chan struct{}, 1)
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
		r             *rpc.Client
		i             = 0
		network, addr string
	)
	routerRing = ketama.NewRing(ketama.Base)
	for serverId, addrs := range Conf.RouterRPCAddrs {
		if network, addr, err = inet.ParseNetwork(addrs); err != nil {
			log.Error("inet.ParseNetwork() error(%v)", err)
			return
		}
		r, err = rpc.Dial(network, addr)
		if err != nil {
			log.Error("rpc.Dial(\"%s\", \"%s\") error(%s)", network, addr, err)
			return
		}
		go rpc.Reconnect(&r, routerQuit, network, addr)
		log.Debug("router rpc addr:%s connect", addr)
		routerServiceMap[serverId] = r
		routerRing.AddNode(serverId, 1)
		i++
	}
	routerRing.Bake()
	return
}

func getRouters() map[string]*rpc.Client {
	return routerServiceMap
}

func RouterClient(userID int64) *rpc.Client {
	node := routerRing.Hash(strconv.FormatInt(userID, 10))
	return routerServiceMap[node]
}

func getRouterNode(userID int64) string {
	return routerRing.Hash(strconv.FormatInt(userID, 10))
}

// divide userIds to corresponding
// response: map[nodes]userIds
func divideNode(userIds []int64) map[string][]int64 {
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

func getSubkeys(serverId string, userIds []int64) (reply *rproto.MGetReply, err error) {
	arg := &rproto.MGetArg{UserIds: userIds}
	reply = &rproto.MGetReply{}
	if err = routerServiceMap[serverId].Call(routerServiceMGet, arg, reply); err != nil {
		log.Error("routerServiceMap[serverId].Call(\"%s\",\"%v\") error(%s)", routerServiceMGet, *arg, err)
	}
	return
}

func getAllSubkeys(serverId string) (reply *rproto.GetAllReply, err error) {
	reply = &rproto.GetAllReply{}
	if err = routerServiceMap[serverId].Call(routerServiceGetAll, nil, reply); err != nil {
		log.Error("routerServiceMap[serverId].Call(\"%s\") error(%s)", routerServiceGetAll, err)
	}
	return
}

package main

import (
	log "code.google.com/p/log4go"
	"github.com/Terry-Mao/goim/libs/hash/ketama"
	rpc "github.com/Terry-Mao/protorpc"
	rproto "github.com/thinkboy/goim/proto/router"
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
	var r *rpc.Client
	routerRing = ketama.NewRing(ketama.Base)
	for i := 0; i < len(Conf.RouterRPCAddrs); i++ {
		r, err = rpc.Dial(Conf.RouterRPCNetworks[i], Conf.RouterRPCAddrs[i])
		if err != nil {
			log.Error("rpc.Dial(\"%s\", \"%s\") error(%s)", Conf.RouterRPCNetworks[i], Conf.RouterRPCAddrs[i], err)
			return
		}
		go rpc.Reconnect(&r, routerQuit, Conf.RouterRPCNetworks[i], Conf.RouterRPCAddrs[i])
		log.Debug("router rpc addr:%s connect", Conf.RouterRPCAddrs[i])
		routerServiceMap[Conf.RouterRPCAddrs[i]] = r
		routerRing.AddNode(Conf.RouterRPCAddrs[i], 1)
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
// response: map[router.addrs]userIds
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

func getSubkeys(routerAddr string, userIds []int64) (reply *rproto.MGetReply, err error) {
	arg := &rproto.MGetArg{UserIds: userIds}
	reply = &rproto.MGetReply{}
	if err = routerServiceMap[routerAddr].Call(routerServiceMGet, arg, reply); err != nil {
		log.Error("routerServiceMap[routerAddr].Call(\"%s\",\"%v\") error(%s)", routerServiceMGet, *arg, err)
	}
	return
}

func getAllSubkeys(routerAddr string) (reply *rproto.GetAllReply, err error) {
	reply = &rproto.GetAllReply{}
	if err = routerServiceMap[routerAddr].Call(routerServiceGetAll, nil, reply); err != nil {
		log.Error("routerServiceMap[routerAddr].Call(\"%s\") error(%s)", routerServiceGetAll, err)
	}
	return
}

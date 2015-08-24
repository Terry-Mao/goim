package main

import (
	"strconv"

	log "code.google.com/p/log4go"
	"github.com/Terry-Mao/goim/define"
	inet "github.com/Terry-Mao/goim/libs/net"
	rproto "github.com/Terry-Mao/goim/proto/router"
	"github.com/Terry-Mao/gopush-cluster/ketama"
	rpc "github.com/Terry-Mao/protorpc"
)

const (
	routerService            = "RouterRPC"
	routerServicePing        = "RouterRPC.Ping"
	routerServiceConnect     = "RouterRPC.Connect"
	routerServiceDisconnect  = "RouterRPC.Disconnect"
	routerServiceGet         = "RouterRPC.Get"
	routerServiceMGet        = "RouterRPC.MGet"
	routerServiceGetSeqCount = "RouterRPC.GetSeqCount"
)

var (
	routerServiceMap = map[string]**rpc.Client{}
	routerRing       *ketama.HashRing
	routerQuit       = make(chan struct{}, 1)
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

/*
func getRouterByUserId(userId int64) (*rpc.Client, error) {
	node := routerRing.Hash(strconv.FormatInt(userId, 10))
	if client, ok := routerServiceMap[node]; !ok || *client == nil {
		return nil, ErrRouter
	} else {
		return *client, nil
	}
}
*/
func getRouterByServer(server string) (*rpc.Client, error) {
	if client, ok := routerServiceMap[server]; !ok || *client == nil {
		return nil, define.ErrRouter
	} else {
		return *client, nil
	}
}

func getRouterNode(userId int64) string {
	return routerRing.Hash(strconv.FormatInt(userId, 10))
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

/*
// 获取在线数量
func getOnlineCount(userID int64) (count int32, err error) {
	c, err := getRouterClient(userID)
	if err != nil {
		log.Error("getRouterClient(\"%d\") error(%v)", userID, err)
		return
	}
	arg := &proto.GetSeqCountArg{UserId: userID}
	reply := &proto.GetSeqCountReply{}
	if err = c.Call(routerServiceGetSeqCount, arg, reply); err != nil {
		log.Error("c.Call(\"%s\",\"%v\") error(%s)", routerServiceGetSeqCount, *arg, err)
		return
	}
	return reply.Count, nil
}
*/

package main

import (
	log "code.google.com/p/log4go"
	"github.com/Terry-Mao/goim/libs/hash/ketama"
	rpc "github.com/Terry-Mao/protorpc"
	"strconv"
	"time"
)

var (
	//routerRPC *RandLB
	routerServiceMap = map[string]*rpc.Client{}
	routerRing       *ketama.HashRing

	routerService           = "RouterRPC"
	routerServicePing       = "RouterRPC.Ping"
	routerServiceConnect    = "RouterRPC.Connect"
	routerServiceDisconnect = "RouterRPC.Disconnect"
)

func InitRouterRpc(addrs []string, retry time.Duration) (err error) {
	var r *rpc.Client
	routerRing = ketama.NewRing(ketama.Base)
	for _, addr := range addrs {
		r, err = rpc.Dial("tcp", addr)
		if err != nil {
			log.Error("rpc.Dial(\"%s\") error(%s)", addr, err)
			return
		}
		go rpcPing(addr, r, retry)
		log.Debug("router rpc addr:%s connect", addr)
		routerServiceMap[addr] = r
		routerRing.AddNode(addr, 1)
	}
	routerRing.Bake()

	return
}

func getRouterClient(userID int64) *rpc.Client {
	node := routerRing.Hash(strconv.FormatInt(userID, 10))
	return routerServiceMap[node]
}

func rpcPing(addr string, c *rpc.Client, retry time.Duration) {
	var err error
	for {
		if err = c.Call(routerServicePing, nil, nil); err != nil {
			log.Error("c.Call(\"%s\", 0, &ret) error(%v), retry after:%ds", routerServicePing, err, retry/time.Second)
			rpcTmp, err := rpc.Dial("tcp", addr)
			if err != nil {
				log.Error("rpc.Dial(\"tcp\", %s) error(%v)", addr, err)
				time.Sleep(retry)
				continue
			}
			c = rpcTmp
			time.Sleep(retry)
			continue
		}
		log.Debug("rpc ping:%s ok", addr)
		time.Sleep(retry)
		continue

	}

}

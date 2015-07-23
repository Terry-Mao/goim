package main

import (
	"strconv"
	"time"

	log "code.google.com/p/log4go"
	"github.com/Terry-Mao/gopush-cluster/ketama"

	"github.com/Terry-Mao/goim/protorpc"
)

var (
	//routerRPC *RandLB
	routerServiceMap = map[string]*protorpc.Client{}
	routerRing       *ketama.HashRing

	routerService           = "RouterRPC"
	routerServicePing       = "RouterRPC.Ping"
	routerServiceConnect    = "RouterRPC.Connect"
	routerServiceDisconnect = "RouterRPC.Disconnect"
)

func InitRouterRpc(addrs []string, retry time.Duration) (err error) {
	var r *protorpc.Client
	routerRing = ketama.NewRing(ketama.Base)
	for _, addr := range addrs {
		r, err = protorpc.Dial("tcp", addr)
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

func getRouterClient(userID int64) *protorpc.Client {
	node := routerRing.Hash(strconv.FormatInt(userID, 10))
	return routerServiceMap[node]
}

func rpcPing(addr string, c *protorpc.Client, retry time.Duration) {
	var err error
	for {
		if err = c.Call(routerServicePing, nil, nil); err != nil {
			log.Error("c.Call(\"%s\", 0, &ret) error(%v), retry after:%ds", routerServicePing, err, retry/time.Second)
			rpcTmp, err := protorpc.Dial("tcp", addr)
			if err != nil {
				log.Error("protorpc.Dial(\"tcp\", %s) error(%v)", addr, err)
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

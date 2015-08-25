package main

import (
	log "code.google.com/p/log4go"
	"github.com/Terry-Mao/protorpc"
	"math/rand"
)

type pushArg struct {
	C       *protorpc.Client
	Server  int32
	SubKeys []string
	Msg     []byte
}

var (
	pushChs []chan *pushArg
)

func InitPush() {
	pushChs = make([]chan *pushArg, Conf.PushChan)
	for i := 0; i < Conf.PushChan; i++ {
		pushChs[i] = make(chan *pushArg, Conf.PushChanSize)
		go processPush(pushChs[i])
	}
}

func processPush(ch chan *pushArg) {
	var arg *pushArg
	for {
		arg = <-ch
		mpushComet(arg.C, arg.Server, arg.SubKeys, arg.Msg)
	}
}

// multi-userids push
func mpush(server int32, subkeys []string, msg []byte) {
	c, err := getCometByServerId(server)
	if err != nil {
		log.Error("getCometByServerId(\"%d\") error(%v)", server, err)
		return
	}
	pushChs[rand.Int()%Conf.PushChan] <- &pushArg{C: c, Server: server, SubKeys: subkeys, Msg: msg}
}

// mssage broadcast
func broadcast(msg []byte) {
	for serverId, c := range cometServiceMap {
		if *c == nil {
			log.Error("broadcast error(%v)", ErrComet)
			return
		}
		// WARN: broadcast called less than mpush, no need a ch for queue
		go broadcastComet(*c, serverId, msg)
	}
}

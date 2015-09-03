package main

import (
	log "code.google.com/p/log4go"
	"github.com/Terry-Mao/goim/define"
	"github.com/Terry-Mao/protorpc"
	"math/rand"
)

type pushArg struct {
	C       *protorpc.Client
	SubKeys []string
	Msg     []byte
	RoomId  int32
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
		if arg.RoomId == define.NoRoom {
			mpushComet(arg.C, arg.SubKeys, arg.Msg)
		} else {
			broadcastRoomComet(arg.C, arg.RoomId, arg.Msg)
		}
	}
}

// multi-userids push
func mpush(server int32, subkeys []string, msg []byte) {
	c, err := getCometByServerId(server)
	if err != nil {
		log.Error("getCometByServerId(\"%d\") error(%v)", server, err)
		return
	}
	pushChs[rand.Int()%Conf.PushChan] <- &pushArg{C: c, SubKeys: subkeys, Msg: msg}
}

// mssage broadcast room
func broadcastRoom(roomId int32, msg []byte) {
	for _, c := range cometServiceMap {
		if *c == nil {
			log.Error("broadcast error(%v)", ErrComet)
			return
		}
		pushChs[rand.Int()%Conf.PushChan] <- &pushArg{C: *c, Msg: msg, RoomId: roomId}
	}
}

// mssage broadcast
func broadcast(msg []byte) {
	for _, c := range cometServiceMap {
		if *c == nil {
			log.Error("broadcast error(%v)", ErrComet)
			return
		}
		// WARN: broadcast called less than mpush, no need a ch for queue
		go broadcastComet(*c, msg)
	}
}

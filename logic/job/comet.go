package main

import (
	log "code.google.com/p/log4go"
	"github.com/Terry-Mao/goim/define"
	inet "github.com/Terry-Mao/goim/libs/net"
	cproto "github.com/Terry-Mao/goim/proto/comet"
	"github.com/Terry-Mao/protorpc"
)

var (
	cometServiceMap = make(map[int32]**protorpc.Client)
)

const (
	CometService              = "PushRPC"
	CometServiceRooms         = "PushRPC.Rooms"
	CometServicePushMsg       = "PushRPC.PushMsg"
	CometServicePushMsgs      = "PushRPC.PushMsgs"
	CometServiceMPushMsg      = "PushRPC.MPushMsg"
	CometServiceMPushMsgs     = "PushRPC.MPushMsgs"
	CometServiceBroadcast     = "PushRPC.Broadcast"
	CometServiceBroadcastRoom = "PushRPC.BroadcastRoom"
)

func InitComet(addrs map[int32]string) (err error) {
	for serverID, addrs := range addrs {
		var (
			rpcClient     *protorpc.Client
			quit          chan struct{}
			network, addr string
		)
		if network, addr, err = inet.ParseNetwork(addrs); err != nil {
			log.Error("inet.ParseNetwork() error(%v)", err)
			return
		}
		if rpcClient, err = protorpc.Dial(network, addr); err != nil {
			log.Error("protorpc.Dial(\"%s\") error(%s)", addr, err)
			return
		}
		go protorpc.Reconnect(&rpcClient, quit, network, addr)
		log.Info("rpc addr:%s connected", addr)
		cometServiceMap[serverID] = &rpcClient
	}
	return
}

// get comet server client by server id
func getCometByServerId(serverID int32) (*protorpc.Client, error) {
	if client, ok := cometServiceMap[serverID]; !ok || *client == nil {
		return nil, ErrComet
	} else {
		return *client, nil
	}
}

func mpushComet(c *protorpc.Client, subkeys []string, body []byte) {
	var (
		args = &cproto.MPushMsgArg{Keys: subkeys, Operation: define.OP_SEND_SMS_REPLY, Msg: body}
		rep  = &cproto.MPushMsgReply{}
		err  error
	)
	if err = c.Call(CometServiceMPushMsg, args, rep); err != nil {
		log.Error("c.Call(\"%s\", %v, reply) error(%v)", CometServiceMPushMsg, *args, err)
	}
}

func broadcastComet(c *protorpc.Client, msg []byte) {
	var (
		args = &cproto.BoardcastArg{Ver: 0, Operation: define.OP_SEND_SMS_REPLY, Msg: msg}
		err  error
	)
	if err = c.Call(CometServiceBroadcast, args, nil); err != nil {
		log.Error("c.Call(\"%s\", %v, reply) error(%v)", CometServiceBroadcast, *args, err)
	}
}

func broadcastRoomComet(c *protorpc.Client, roomId int32, msg []byte) {
	var (
		args = &cproto.BoardcastRoomArg{Ver: 0, Operation: define.OP_SEND_SMS_REPLY, Msg: msg, RoomId: roomId}
		err  error
	)
	if err = c.Call(CometServiceBroadcastRoom, args, nil); err != nil {
		log.Error("c.Call(\"%s\", %v, reply) error(%v)", CometServiceBroadcastRoom, *args, err)
	}
}

func roomsComet(c *protorpc.Client) map[int32]bool {
	var (
		reply = &cproto.RoomsReply{}
		err   error
	)
	if err = c.Call(CometServiceRooms, nil, reply); err != nil {
		log.Error("c.Call(\"%s\", nil, reply) error(%v)", CometServiceRooms, err)
		return nil
	}
	return reply.Rooms
}

package main

import (
	"net/rpc"
	"time"

	log "code.google.com/p/log4go"
	"github.com/Terry-Mao/goim/libs/define"
	inet "github.com/Terry-Mao/goim/libs/net"
	cproto "github.com/thinkboy/goim/libs/proto/comet"
)

var (
	cometServiceMap = make(map[int32]**rpc.Client)
)

const (
	CometService              = "PushRPC"
	CometServicePing          = "PushRPC.Ping"
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
			rpcClient     *rpc.Client
			quit          chan struct{}
			network, addr string
		)
		if network, addr, err = inet.ParseNetwork(addrs); err != nil {
			log.Error("inet.ParseNetwork() error(%v)", err)
			return
		}
		if rpcClient, err = rpc.Dial(network, addr); err != nil {
			log.Error("rpc.Dial(\"%s\") error(%s)", addr, err)
			return
		}
		go Reconnect(&rpcClient, quit, network, addr)
		log.Info("rpc addr:%s connected", addr)
		cometServiceMap[serverID] = &rpcClient
	}
	return
}

// Reconnect for ping rpc server and reconnect with it when it's crash.
func Reconnect(dst **rpc.Client, quit chan struct{}, network, address string) {
	var (
		tmp    *rpc.Client
		err    error
		call   *rpc.Call
		ch     = make(chan *rpc.Call, 1)
		client = *dst
	)
	for {
		select {
		case <-quit:
			return
		default:
			if client != nil {
				call = <-client.Go(CometServicePing, 0, 0, ch).Done
			}
			if client == nil || call.Error != nil {
				if tmp, err = rpc.Dial(network, address); err == nil {
					*dst = tmp
					client = tmp
				}
			}
		}
		time.Sleep(1 * time.Second)
	}
}

// get comet server client by server id
func getCometByServerId(serverID int32) (*rpc.Client, error) {
	if client, ok := cometServiceMap[serverID]; !ok || *client == nil {
		return nil, ErrComet
	} else {
		return *client, nil
	}
}

func mpushComet(c *rpc.Client, subkeys []string, body []byte) {
	var (
		args = &cproto.MPushMsgArg{Keys: subkeys, Ver: 0, Operation: define.OP_SEND_SMS_REPLY, Msg: body}
		rep  = &cproto.MPushMsgReply{}
		err  error
	)
	if err = c.Call(CometServiceMPushMsg, args, rep); err != nil {
		log.Error("c.Call(\"%s\", %v, reply) error(%v)", CometServiceMPushMsg, *args, err)
	}
}

func broadcastComet(c *rpc.Client, msg []byte) {
	var (
		args = &cproto.BoardcastArg{Ver: 0, Operation: define.OP_SEND_SMS_REPLY, Msg: msg}
		err  error
	)
	if err = c.Call(CometServiceBroadcast, args, 0); err != nil {
		log.Error("c.Call(\"%s\", %v, reply) error(%v)", CometServiceBroadcast, *args, err)
	}
}

func broadcastRoomComet(c *rpc.Client, roomId int32, msg []byte) {
	var (
		args = &cproto.BoardcastRoomArg{Ver: 0, Operation: define.OP_SEND_SMS_REPLY, Msg: msg, RoomId: roomId}
		err  error
	)
	if err = c.Call(CometServiceBroadcastRoom, args, 0); err != nil {
		log.Error("c.Call(\"%s\", %v, reply) error(%v)", CometServiceBroadcastRoom, *args, err)
	}
}

func roomsComet(c *rpc.Client) map[int32]bool {
	var (
		reply = &cproto.RoomsReply{}
		err   error
	)
	if err = c.Call(CometServiceRooms, 0, reply); err != nil {
		log.Error("c.Call(\"%s\", nil, reply) error(%v)", CometServiceRooms, err)
		return nil
	}
	return reply.Rooms
}

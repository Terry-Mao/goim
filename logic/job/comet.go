package main

import (
	"encoding/json"
	"net/rpc"
	"time"

	log "code.google.com/p/log4go"
	"github.com/Terry-Mao/goim/libs/define"
	inet "github.com/Terry-Mao/goim/libs/net"
	"github.com/Terry-Mao/goim/libs/proto"
)

var (
	cometServiceMap = make(map[int32]**rpc.Client)
)

const (
	CometService              = "PushRPC"
	CometServicePing          = "PushRPC.Ping"
	CometServiceRooms         = "PushRPC.Rooms"
	CometServicePushMsg       = "PushRPC.PushMsg"
	CometServiceMPushMsg      = "PushRPC.MPushMsg"
	CometServiceBroadcast     = "PushRPC.Broadcast"
	CometServiceBroadcastRoom = "PushRPC.BroadcastRoom"
)

func InitComet(addrs map[int32]string) (err error) {
	for serverID, addrsTmp := range addrs {
		var (
			rpcClient     *rpc.Client
			quit          chan struct{}
			network, addr string
		)
		if network, addr, err = inet.ParseNetwork(addrsTmp); err != nil {
			log.Error("inet.ParseNetwork() error(%v)", err)
			return
		}
		if rpcClient, err = rpc.Dial(network, addr); err != nil {
			log.Error("rpc.Dial(\"%s\") error(%s)", addr, err)
		}
		go Reconnect(&rpcClient, quit, network, addr)
		log.Info("init comet rpc addr:%s connection", addr)
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
		args   = proto.NoArg{}
		reply  = proto.NoReply{}
	)
	for {
		select {
		case <-quit:
			return
		default:
			if client != nil {
				call = <-client.Go(CometServicePing, &args, &reply, ch).Done
				if call.Error != nil {
					log.Error("rpc ping %s error(%v)", address, call.Error)
				}
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
func getCometByServerId(serverId int32) (*rpc.Client, error) {
	if client, ok := cometServiceMap[serverId]; !ok || *client == nil {
		return nil, ErrComet
	} else {
		return *client, nil
	}
}

func mPushComet(serverId int32, subkeys []string, body json.RawMessage) {
	var (
		p     = &proto.Proto{Ver: 0, Operation: define.OP_SEND_SMS_REPLY, Body: body}
		args  = &proto.MPushMsgArg{Keys: subkeys, P: p}
		reply = &proto.MPushMsgReply{}
		c     *rpc.Client
		err   error
	)
	c, err = getCometByServerId(serverId)
	if err != nil {
		log.Error("getCometByServerId(\"%d\") error(%v)", serverId, err)
		return
	}
	if err = c.Call(CometServiceMPushMsg, args, reply); err != nil {
		log.Error("c.Call(\"%s\", %v, reply) error(%v)", CometServiceMPushMsg, *args, err)
	}
}

func broadcast(msg []byte) {
	var (
		p    = &proto.Proto{Ver: 0, Operation: define.OP_SEND_SMS_REPLY, Body: msg}
		args = &proto.BoardcastArg{P: p}
	)
	for serverId, c := range cometServiceMap {
		if *c != nil {
			go broadcastComet(*c, args)
		} else {
			log.Error("doesn`t push message to serverId:%d", serverId)
		}
	}
}

func broadcastComet(c *rpc.Client, args *proto.BoardcastArg) (err error) {
	var reply = proto.NoReply{}
	if err = c.Call(CometServiceBroadcast, args, &reply); err != nil {
		log.Error("c.Call(\"%s\", %v, reply) error(%v)", CometServiceBroadcast, *args, err)
	}
	return
}

func broadcastRoomBytes(roomId int32, body []byte) {
	var (
		p        = &proto.Proto{Ver: 0, Operation: define.OP_RAW, Body: body}
		args     = proto.BoardcastRoomArg{P: p, RoomId: roomId}
		reply    = proto.NoReply{}
		c        *rpc.Client
		serverId int32
		servers  map[int32]struct{}
		err      error
		ok       bool
	)

	if servers, ok = RoomServersMap[roomId]; ok {
		for serverId, _ = range servers {
			if c, err = getCometByServerId(serverId); err != nil {
				log.Error("getCometByServerId(%d) error(%v)", serverId, err)
				continue
			}
			if err = c.Call(CometServiceBroadcastRoom, &args, &reply); err != nil {
				log.Error("c.Call(\"%s\", %v, reply) serverId:%d error(%v)", CometServiceBroadcastRoom, args, serverId, err)
			}
		}
	}
}

func roomsComet(c *rpc.Client) []int32 {
	var (
		args  = proto.NoArg{}
		reply = proto.RoomsReply{}
		err   error
	)
	if err = c.Call(CometServiceRooms, &args, &reply); err != nil {
		log.Error("c.Call(\"%s\", 0, reply) error(%v)", CometServiceRooms, err)
		return nil
	}
	return reply.RoomIds
}

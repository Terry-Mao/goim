package main

import (
	"encoding/json"
	"goim/libs/define"
	inet "goim/libs/net"
	"goim/libs/proto"
	"net/rpc"
	"time"

	log "github.com/thinkboy/log4go"
	"sync/atomic"
)

var (
	cometServiceMap = make(map[int32]*Comet)
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

type Comet struct {
	serverId      int32
	rpcClient     **rpc.Client
	routines      []chan *proto.BoardcastRoomArg
	routinesNum   int64
	routineAmount int64
	routineSize   int
}

func (cm *Comet) BroadcastRoom(arg *proto.BoardcastRoomArg) (err error) {
	num := atomic.AddInt64(&cm.routinesNum, 1) % cm.routineAmount
	select {
	case cm.routines[num] <- arg:
	default:
		err = ErrCometFull
	}
	return
}

// room process
func (cm *Comet) roomproc(c chan *proto.BoardcastRoomArg) {
	var (
		arg       *proto.BoardcastRoomArg
		reply     = &proto.NoReply{}
		rpcClient *rpc.Client
		err       error
	)
	for {
		arg = <-c
		// room push
		if rpcClient, err = getCometByServerId(cm.serverId); err != nil {
			log.Error("getCometByServerId(\"%d\") error(%v)", cm.serverId, err)
			continue
		}
		if err = rpcClient.Call(CometServiceBroadcastRoom, arg, reply); err != nil {
			log.Error("c.Call(\"%s\", %v, reply) serverId:%d error(%v)", CometServiceBroadcastRoom, arg, cm.serverId, err)
		}
		arg = nil
		err = nil
	}
}

func InitComet(addrs map[int32]string, routineAmount int64, routineSize int) (err error) {
	for serverID, addrsTmp := range addrs {
		var (
			cm            *Comet
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
		// comet
		cm = new(Comet)
		cm.serverId = serverID
		cm.routines = make([]chan *proto.BoardcastRoomArg, routineAmount)
		cm.routineAmount = routineAmount
		cm.routineSize = routineSize
		cm.rpcClient = &rpcClient
		cometServiceMap[serverID] = cm
		// process
		for i := int64(0); i < routineAmount; i++ {
			c := make(chan *proto.BoardcastRoomArg, routineSize)
			cm.routines[i] = c
			go cm.roomproc(c)
		}
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
	if cm, ok := cometServiceMap[serverId]; !ok || cm.routines == nil {
		return nil, ErrComet
	} else {
		return *cm.rpcClient, nil
	}
}

func getClientByComet(cm *Comet) (*rpc.Client, error) {
	if cm.rpcClient != nil && *cm.rpcClient != nil {
		return *cm.rpcClient, nil
	} else {
		return nil, ErrComet
	}
}

// mPushComet push a message to a batch of subkeys
func mPushComet(serverId int32, subkeys []string, body json.RawMessage) {
	var (
		args = &proto.MPushMsgArg{Keys: subkeys,
			P: proto.Proto{Ver: 0, Operation: define.OP_SEND_SMS_REPLY, Body: body, Time: time.Now()}}
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

// broadcast broadcast a message to all
func broadcast(msg []byte) {
	var (
		args      = &proto.BoardcastArg{P: proto.Proto{Ver: 0, Operation: define.OP_SEND_SMS_REPLY, Body: msg, Time: time.Now()}}
		rpcClient *rpc.Client
		err       error
	)
	for serverId, cm := range cometServiceMap {
		if rpcClient, err = getClientByComet(cm); err != nil {
			log.Error("getClientByComet(\"%d\") error(%v)", serverId, err)
			continue
		}
		go broadcastComet(rpcClient, args)
	}
}

// broadcastComet a message to specified comet
func broadcastComet(c *rpc.Client, args *proto.BoardcastArg) (err error) {
	var reply = proto.NoReply{}
	if err = c.Call(CometServiceBroadcast, args, &reply); err != nil {
		log.Error("c.Call(\"%s\", %v, reply) error(%v)", CometServiceBroadcast, *args, err)
	}
	return
}

// broadcastRoomBytes broadcast aggregation messages to room
func broadcastRoomBytes(roomId int32, body []byte) {
	var (
		args     = proto.BoardcastRoomArg{P: proto.Proto{Ver: 0, Operation: define.OP_RAW, Body: body, Time: time.Now()}, RoomId: roomId}
		cm       *Comet
		serverId int32
		servers  map[int32]struct{}
		ok       bool
		err      error
	)
	if servers, ok = RoomServersMap[roomId]; ok {
		for serverId, _ = range servers {
			if cm, ok = cometServiceMap[serverId]; ok {
				// push routines
				if err = cm.BroadcastRoom(&args); err != nil {
					log.Error("broadcastRoomBytes roomId:%d error(%v)", roomId, err)
				}
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

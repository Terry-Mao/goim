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

type CometOptions struct {
	RoutineAmount int64
	RoutineSize   int
	CallSize      int
}

type Comet struct {
	serverId             int32
	rpcClient            *rpc.Client
	pushRoutines         []chan *proto.MPushMsgArg
	broadcastRoutines    []chan *proto.BoardcastArg
	roomRoutines         []chan *proto.BoardcastRoomArg
	pushRoutinesNum      int64
	roomRoutinesNum      int64
	broadcastRoutinesNum int64
	options              CometOptions
}

// user push
func (cm *Comet) Push(arg *proto.MPushMsgArg) (err error) {
	num := atomic.AddInt64(&cm.pushRoutinesNum, 1) % cm.options.RoutineAmount
	cm.pushRoutines[num] <- arg
	return
}

// room push
func (cm *Comet) BroadcastRoom(arg *proto.BoardcastRoomArg) (err error) {
	num := atomic.AddInt64(&cm.roomRoutinesNum, 1) % cm.options.RoutineAmount
	cm.roomRoutines[num] <- arg
	return
}

// broadcast
func (cm *Comet) Broadcast(arg *proto.BoardcastArg) (err error) {
	num := atomic.AddInt64(&cm.broadcastRoutinesNum, 1) % cm.options.RoutineAmount
	cm.broadcastRoutines[num] <- arg
	return
}

// process
func (c *Comet) process(pushChan chan *proto.MPushMsgArg, roomChan chan *proto.BoardcastRoomArg, broadcastChan chan *proto.BoardcastArg) {
	var (
		pushArg      *proto.MPushMsgArg
		roomArg      *proto.BoardcastRoomArg
		broadcastArg *proto.BoardcastArg
		reply        = &proto.NoReply{}
		done         = make(chan *rpc.Call, c.options.CallSize)
		call         *rpc.Call
	)
	for {
		select {
		case pushArg = <-pushChan:
			// push
			if c.rpcClient != nil {
				c.rpcClient.Go(CometServiceMPushMsg, pushArg, reply, done)
			} else {
				log.Error("rpcClient.Go(%s, %v, reply, done) serverId:%d error(%v)", CometServiceMPushMsg, pushArg, c.serverId, ErrComet)
			}
			pushArg = nil
		case roomArg = <-roomChan:
			// room
			if c.rpcClient != nil {
				c.rpcClient.Go(CometServiceBroadcastRoom, roomArg, reply, done)
			} else {
				log.Error("rpcClient.Go(%s, %v, reply, done) serverId:%d error(%v)", CometServiceBroadcastRoom, roomArg, c.serverId, ErrComet)
			}
			roomArg = nil
		case broadcastArg = <-broadcastChan:
			// broadcast
			if c.rpcClient != nil {
				c.rpcClient.Go(CometServiceBroadcast, broadcastArg, reply, done)
			} else {
				log.Error("rpcClient.Go(%s, %v, reply, done) serverId:%d error(%v)", CometServiceBroadcast, broadcastArg, c.serverId, ErrComet)
			}
			broadcastArg = nil
		case call = <-done:
			// result
			if call.Error != nil {
				log.Error("rpcClient.Go(%s, %v, reply, done) serverId:%d error(%v)", call.ServiceMethod, call.Args, c.serverId, call.Error)
			}
			call = nil
		}
	}
}

// Reconnect for ping rpc server and reconnect with it when it's crash.
func (c *Comet) ping(network, address string) {
	var (
		call  *rpc.Call
		ch    = make(chan *rpc.Call, 1)
		args  = proto.NoArg{}
		reply = proto.NoReply{}
	)
	for {
		if c.rpcClient != nil {
			call = <-c.rpcClient.Go(CometServicePing, &args, &reply, ch).Done
			if call.Error != nil {
				log.Error("rpc ping %s error(%v)", address, call.Error)
			}
		}
		if c.rpcClient == nil || call.Error != nil {
			if newCli, err := rpc.Dial(network, address); err == nil {
				c.rpcClient = newCli
			}
		}
		time.Sleep(1 * time.Second)
	}
}

func InitComet(addrs map[int32]string, options CometOptions) (err error) {
	for serverID, addrsTmp := range addrs {
		var (
			c             *Comet
			rpcClient     *rpc.Client
			network, addr string
		)
		if network, addr, err = inet.ParseNetwork(addrsTmp); err != nil {
			log.Error("inet.ParseNetwork() error(%v)", err)
			return
		}
		if rpcClient, err = rpc.Dial(network, addr); err != nil {
			log.Error("rpc.Dial(\"%s\") error(%s)", addr, err)
		}
		// comet
		c = new(Comet)
		c.serverId = serverID
		c.pushRoutines = make([]chan *proto.MPushMsgArg, options.RoutineAmount)
		c.roomRoutines = make([]chan *proto.BoardcastRoomArg, options.RoutineAmount)
		c.broadcastRoutines = make([]chan *proto.BoardcastArg, options.RoutineAmount)
		c.options = options
		c.rpcClient = rpcClient
		cometServiceMap[serverID] = c
		// process
		for i := int64(0); i < options.RoutineAmount; i++ {
			pushChan := make(chan *proto.MPushMsgArg, options.RoutineSize)
			roomChan := make(chan *proto.BoardcastRoomArg, options.RoutineSize)
			broadcastChan := make(chan *proto.BoardcastArg, options.RoutineSize)
			c.pushRoutines[i] = pushChan
			c.roomRoutines[i] = roomChan
			c.broadcastRoutines[i] = broadcastChan
			go c.process(pushChan, roomChan, broadcastChan)
		}
		// ping & reconnect
		go c.ping(network, addr)
		log.Info("init comet rpc addr:%s connection", addr)
	}
	return
}

// mPushComet push a message to a batch of subkeys
func mPushComet(serverId int32, subKeys []string, body json.RawMessage) {
	var args = &proto.MPushMsgArg{
		Keys: subKeys, P: proto.Proto{Ver: 0, Operation: define.OP_SEND_SMS_REPLY, Body: body, Time: time.Now()},
	}
	if c, ok := cometServiceMap[serverId]; ok {
		if err := c.Push(args); err != nil {
			log.Error("c.Push(%v) serverId:%d error(%v)", args, serverId, err)
		}
	}
}

// broadcast broadcast a message to all
func broadcast(msg []byte) {
	var args = &proto.BoardcastArg{
		P: proto.Proto{Ver: 0, Operation: define.OP_SEND_SMS_REPLY, Body: msg, Time: time.Now()},
	}
	for serverId, c := range cometServiceMap {
		if err := c.Broadcast(args); err != nil {
			log.Error("c.Broadcast(%v) serverId:%d error(%v)", args, serverId, err)
		}
	}
}

// broadcastRoomBytes broadcast aggregation messages to room
func broadcastRoomBytes(roomId int32, body []byte) {
	var (
		args     = proto.BoardcastRoomArg{P: proto.Proto{Ver: 0, Operation: define.OP_RAW, Body: body, Time: time.Now()}, RoomId: roomId}
		c        *Comet
		serverId int32
		servers  map[int32]struct{}
		ok       bool
		err      error
	)
	if servers, ok = RoomServersMap[roomId]; ok {
		for serverId, _ = range servers {
			if c, ok = cometServiceMap[serverId]; ok {
				// push routines
				if err = c.BroadcastRoom(&args); err != nil {
					log.Error("c.BroadcastRoom(%v) roomId:%d error(%v)", args, roomId, err)
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
		log.Error("c.Call(%s, 0, reply) error(%v)", CometServiceRooms, err)
		return nil
	}
	return reply.RoomIds
}

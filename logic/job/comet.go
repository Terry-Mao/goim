package main

import (
	"encoding/json"
	"goim/libs/define"
	inet "goim/libs/net"
	"goim/libs/net/xrpc"
	"goim/libs/proto"

	log "github.com/thinkboy/log4go"
	"strings"
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
	RoutineSize uint64
	RoutineChan int
}

type Comet struct {
	serverId             int32
	rpcClient            *xrpc.Clients
	pushRoutines         []chan *proto.MPushMsgArg
	broadcastRoutines    []chan *proto.BoardcastArg
	roomRoutines         []chan *proto.BoardcastRoomArg
	pushRoutinesNum      uint64
	roomRoutinesNum      uint64
	broadcastRoutinesNum uint64
	options              CometOptions
}

// user push
func (c *Comet) Push(arg *proto.MPushMsgArg) (err error) {
	num := atomic.AddUint64(&c.pushRoutinesNum, 1) % c.options.RoutineSize
	c.pushRoutines[num] <- arg
	return
}

// room push
func (c *Comet) BroadcastRoom(arg *proto.BoardcastRoomArg) (err error) {
	num := atomic.AddUint64(&c.roomRoutinesNum, 1) % c.options.RoutineSize
	c.roomRoutines[num] <- arg
	return
}

// broadcast
func (c *Comet) Broadcast(arg *proto.BoardcastArg) (err error) {
	num := atomic.AddUint64(&c.broadcastRoutinesNum, 1) % c.options.RoutineSize
	c.broadcastRoutines[num] <- arg
	return
}

// process
func (c *Comet) process(pushChan chan *proto.MPushMsgArg, roomChan chan *proto.BoardcastRoomArg, broadcastChan chan *proto.BoardcastArg) {
	var (
		pushArg      *proto.MPushMsgArg
		roomArg      *proto.BoardcastRoomArg
		broadcastArg *proto.BoardcastArg
		reply        = &proto.NoReply{}
		err          error
	)
	for {
		select {
		case pushArg = <-pushChan:
			// push
			err = c.rpcClient.Call(CometServiceMPushMsg, pushArg, reply)
			if err != nil {
				log.Error("rpcClient.Call(%s, %v, reply) serverId:%d error(%v)", CometServiceMPushMsg, pushArg, c.serverId, err)
				DefaultStat.IncrPushMsgFailed()
			}
			pushArg = nil
		case roomArg = <-roomChan:
			// room
			err = c.rpcClient.Call(CometServiceBroadcastRoom, roomArg, reply)
			if err != nil {
				log.Error("rpcClient.Call(%s, %v, reply) serverId:%d error(%v)", CometServiceBroadcastRoom, roomArg, c.serverId, err)
				DefaultStat.IncrBroadcastRoomMsgFailed()
			}
			roomArg = nil
		case broadcastArg = <-broadcastChan:
			// broadcast
			err = c.rpcClient.Call(CometServiceBroadcast, broadcastArg, reply)
			if err != nil {
				log.Error("rpcClient.Call(%s, %v, reply) serverId:%d error(%v)", CometServiceBroadcast, broadcastArg, c.serverId, err)
				DefaultStat.IncrBroadcastMsgFailed()
			}
			broadcastArg = nil
		}
	}
}

func InitComet(addrs map[int32]string, options CometOptions) (err error) {
	var (
		serverId      int32
		bind          string
		network, addr string
	)
	for serverId, bind = range addrs {
		var rpcOptions []xrpc.ClientOptions
		for _, bind = range strings.Split(bind, ",") {
			if network, addr, err = inet.ParseNetwork(bind); err != nil {
				log.Error("inet.ParseNetwork() error(%v)", err)
				return
			}
			options := xrpc.ClientOptions{
				Proto: network,
				Addr:  addr,
			}
			rpcOptions = append(rpcOptions, options)
		}
		// rpc clients
		rpcClient := xrpc.Dials(rpcOptions)
		// ping & reconnect
		rpcClient.Ping(CometServicePing)
		// comet
		c := new(Comet)
		c.serverId = serverId
		c.rpcClient = rpcClient
		c.pushRoutines = make([]chan *proto.MPushMsgArg, options.RoutineSize)
		c.roomRoutines = make([]chan *proto.BoardcastRoomArg, options.RoutineSize)
		c.broadcastRoutines = make([]chan *proto.BoardcastArg, options.RoutineSize)
		c.options = options
		cometServiceMap[serverId] = c
		// process
		for i := uint64(0); i < options.RoutineSize; i++ {
			pushChan := make(chan *proto.MPushMsgArg, options.RoutineChan)
			roomChan := make(chan *proto.BoardcastRoomArg, options.RoutineChan)
			broadcastChan := make(chan *proto.BoardcastArg, options.RoutineChan)
			c.pushRoutines[i] = pushChan
			c.roomRoutines[i] = roomChan
			c.broadcastRoutines[i] = broadcastChan
			go c.process(pushChan, roomChan, broadcastChan)
		}
		log.Info("init comet rpc: %v", rpcOptions)
	}
	return
}

// mPushComet push a message to a batch of subkeys
func mPushComet(serverId int32, subKeys []string, body json.RawMessage) {
	var args = proto.MPushMsgArg{
		Keys: subKeys, P: proto.Proto{Ver: 0, Operation: define.OP_SEND_SMS_REPLY, Body: body},
	}
	if c, ok := cometServiceMap[serverId]; ok {
		if err := c.Push(&args); err != nil {
			log.Error("c.Push(%v) serverId:%d error(%v)", args, serverId, err)
			DefaultStat.IncrPushMsgFailed()
		}
	}
	DefaultStat.IncrPushMsg()
}

// broadcast broadcast a message to all
func broadcast(msg []byte) {
	var args = proto.BoardcastArg{
		P: proto.Proto{Ver: 0, Operation: define.OP_SEND_SMS_REPLY, Body: msg},
	}
	for serverId, c := range cometServiceMap {
		if err := c.Broadcast(&args); err != nil {
			log.Error("c.Broadcast(%v) serverId:%d error(%v)", args, serverId, err)
			DefaultStat.IncrBroadcastMsgFailed()
		}
	}
	DefaultStat.IncrBroadcastMsg()
}

// broadcastRoomBytes broadcast aggregation messages to room
func broadcastRoomBytes(roomId int32, body []byte) {
	var (
		args     = proto.BoardcastRoomArg{P: proto.Proto{Ver: 0, Operation: define.OP_RAW, Body: body}, RoomId: roomId}
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
					DefaultStat.IncrBroadcastRoomMsgFailed()
				}
			}
		}
	}
	DefaultStat.IncrBroadcastRoomMsg()
}

func roomsComet(c *xrpc.Clients) map[int32]struct{} {
	var (
		args  = proto.NoArg{}
		reply = proto.RoomsReply{}
		err   error
	)
	if err = c.Call(CometServiceRooms, &args, &reply); err != nil {
		log.Error("c.Call(%s, args, reply) error(%v)", CometServiceRooms, err)
		return nil
	}
	return reply.RoomIds
}

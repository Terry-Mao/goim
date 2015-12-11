package main

import (
	log "code.google.com/p/log4go"
	inet "github.com/Terry-Mao/goim/libs/net"
	proto "github.com/Terry-Mao/goim/libs/proto/comet"
	rpc "github.com/Terry-Mao/protorpc"
	"net"
)

func InitRPCPush(addrs []string) (err error) {
	var (
		bind          string
		network, addr string
		c             = &PushRPC{}
	)
	rpc.Register(c)
	for _, bind = range addrs {
		if network, addr, err = inet.ParseNetwork(bind); err != nil {
			log.Error("inet.ParseNetwork() error(%v)", err)
			return
		}
		go rpcListen(network, addr)
	}
	return
}

func rpcListen(network, addr string) {
	l, err := net.Listen(network, addr)
	if err != nil {
		log.Error("net.Listen(\"%s\", \"%s\") error(%v)", network, addr, err)
		panic(err)
	}
	// if process exit, then close the rpc addr
	defer func() {
		log.Info("listen rpc: \"%s\" close", addr)
		if err := l.Close(); err != nil {
			log.Error("listener.Close() error(%v)", err)
		}
	}()
	rpc.Accept(l)
}

// Push RPC
type PushRPC struct {
}

// Push push a message to a specified sub key
func (this *PushRPC) PushMsg(arg *proto.PushMsgArg, reply *proto.NoReply) (err error) {
	var (
		bucket  *Bucket
		channel *Channel
	)
	if arg == nil {
		err = ErrPushMsgArg
		return
	}
	bucket = DefaultServer.Bucket(arg.Key)
	if channel = bucket.Channel(arg.Key); channel != nil {
		err = channel.PushMsg(int16(arg.Ver), arg.Operation, arg.Msg)
	}
	return
}

// Pushs push multiple messages to a specified sub key
func (this *PushRPC) PushMsgs(arg *proto.PushMsgsArg, reply *proto.PushMsgsReply) (err error) {
	var (
		bucket  *Bucket
		channel *Channel
	)
	reply.Index = -1
	if arg == nil || len(arg.Vers) != len(arg.Operations) || len(arg.Operations) != len(arg.Msgs) {
		err = ErrPushMsgsArg
		return
	}
	bucket = DefaultServer.Bucket(arg.Key)
	if channel = bucket.Channel(arg.Key); channel != nil {
		reply.Index, err = channel.PushMsgs(arg.Vers, arg.Operations, arg.Msgs)
	}
	return
}

// Push push a message to a specified sub key
func (this *PushRPC) MPushMsg(arg *proto.MPushMsgArg, reply *proto.MPushMsgReply) (err error) {
	var (
		bucket  *Bucket
		channel *Channel
		key     string
		n       int
	)
	reply.Index = -1
	if arg == nil {
		err = ErrMPushMsgArg
		return
	}
	for n, key = range arg.Keys {
		bucket = DefaultServer.Bucket(key)
		if channel = bucket.Channel(key); channel != nil {
			if err = channel.PushMsg(int16(arg.Ver), arg.Operation, arg.Msg); err != nil {
				return
			}
			reply.Index = int32(n)
		}
	}
	return
}

// MPushMsgs push msgs to multiple user.
func (this *PushRPC) MPushMsgs(arg *proto.MPushMsgsArg, reply *proto.MPushMsgsReply) (err error) {
	var (
		bucket  *Bucket
		channel *Channel
		key     string
		n       int
	)
	reply.Index = -1
	if arg == nil || len(arg.Keys) != len(arg.Vers) || len(arg.Vers) != len(arg.Operations) || len(arg.Operations) != len(arg.Msgs) {
		err = ErrMPushMsgsArg
		return
	}
	for n, key = range arg.Keys {
		bucket = DefaultServer.Bucket(key)
		if channel = bucket.Channel(key); channel != nil {
			if err = channel.PushMsg(int16(arg.Vers[n]), arg.Operations[n], arg.Msgs[n]); err != nil {
				return
			}
			reply.Index = int32(n)
		}
	}
	return
}

func (this *PushRPC) Broadcast(arg *proto.BoardcastArg, reply *proto.NoReply) (err error) {
	var bucket *Bucket
	for _, bucket = range DefaultServer.Buckets {
		go bucket.Broadcast(int16(arg.Ver), arg.Operation, arg.Msg)
	}
	return
}

func (this *PushRPC) BroadcastRoom(arg *proto.BoardcastRoomArg, reply *proto.NoReply) (err error) {
	var (
		bucket *Bucket
		room   *Room
	)
	for _, bucket = range DefaultServer.Buckets {
		if room = bucket.Room(arg.RoomId); room != nil {
			go room.PushMsg(int16(arg.Ver), arg.Operation, arg.Msg)
		}
	}
	return
}

func (this *PushRPC) Rooms(arg *proto.NoArgs, reply *proto.RoomsReply) (err error) {
	var (
		roomId int32
		bucket *Bucket
		rooms  = make(map[int32]bool)
	)
	for _, bucket = range DefaultServer.Buckets {
		for roomId, _ = range bucket.Rooms() {
			rooms[roomId] = true
		}
	}
	reply.Rooms = rooms
	return
}

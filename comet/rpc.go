package main

import (
	inet "goim/libs/net"
	"goim/libs/proto"
	"net"
	"net/rpc"

	log "github.com/thinkboy/log4go"
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

func (this *PushRPC) Ping(arg *proto.NoArg, reply *proto.NoReply) error {
	return nil
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
		err = channel.Push(&arg.P)
	}
	return
}

// Push push a message to specified sub keys
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
			if err = channel.Push(&arg.P); err != nil {
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
		n       int32
		PMArg   *proto.PushMsgArg
	)
	reply.Index = -1
	if arg == nil {
		err = ErrMPushMsgsArg
		return
	}
	for _, PMArg = range arg.PMArgs {
		bucket = DefaultServer.Bucket(PMArg.Key)
		if channel = bucket.Channel(PMArg.Key); channel != nil {
			if err = channel.Push(&PMArg.P); err != nil {
				return
			}
			n++
			reply.Index = n
		}
	}
	return
}

// Broadcast broadcast msg to all user.
func (this *PushRPC) Broadcast(arg *proto.BoardcastArg, reply *proto.NoReply) (err error) {
	var bucket *Bucket
	for _, bucket = range DefaultServer.Buckets {
		go bucket.Broadcast(&arg.P)
	}
	return
}

// Broadcast broadcast msg to specified room.
func (this *PushRPC) BroadcastRoom(arg *proto.BoardcastRoomArg, reply *proto.NoReply) (err error) {
	var bucket *Bucket
	for _, bucket = range DefaultServer.Buckets {
		bucket.BroadcastRoom(arg)
	}
	return
}

func (this *PushRPC) Rooms(arg *proto.NoArg, reply *proto.RoomsReply) (err error) {
	var (
		roomId  int32
		bucket  *Bucket
		roomIds = make(map[int32]struct{})
	)
	for _, bucket = range DefaultServer.Buckets {
		for roomId, _ = range bucket.Rooms() {
			roomIds[roomId] = struct{}{}
		}
	}
	reply.RoomIds = roomIds
	return
}

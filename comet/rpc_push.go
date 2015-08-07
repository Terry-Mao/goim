package main

import (
	log "code.google.com/p/log4go"
	proto "github.com/Terry-Mao/goim/proto/comet"
	rpc "github.com/Terry-Mao/protorpc"
	"net"
)

func InitRPCPush() error {
	c := &PushRPC{}
	rpc.Register(c)
	for _, bind := range Conf.RPCPushBind {
		log.Info("start listen rpc addr: \"%s\"", bind)
		go rpcListen(Conf.RPCPushNetwork, bind)
	}
	return nil
}

func rpcListen(network, addr string) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Error("net.Listen(\"tcp\", \"%s\") error(%v)", addr, err)
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

// Push push a message to a specified sub key, must goroutine safe.
func (this *PushRPC) PushMsg(arg *proto.PushMsgArg, reply *proto.NoReply) (err error) {
	if arg == nil {
		err = ErrPushMsgArg
		return
	}
	bucket := DefaultServer.Bucket(arg.Key)
	if channel := bucket.Get(arg.Key); channel != nil {
		err = channel.PushMsg(int16(arg.Ver), arg.Operation, arg.Msg)
	}
	return
}

// Pushs push multiple messages to a specified sub key, must goroutine safe.
func (this *PushRPC) PushMsgs(arg *proto.PushMsgsArg, reply *proto.PushMsgsReply) (err error) {
	reply.Index = -1
	if arg == nil || len(arg.Vers) != len(arg.Operations) || len(arg.Operations) != len(arg.Msgs) {
		err = ErrPushMsgsArg
		return
	}
	bucket := DefaultServer.Bucket(arg.Key)
	if channel := bucket.Get(arg.Key); channel != nil {
		reply.Index, err = channel.PushMsgs(arg.Vers, arg.Operations, arg.Msgs)
	}
	return
}

// Push push a message to a specified sub key, must goroutine safe.
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
		if channel = bucket.Get(key); channel != nil {
			if err = channel.PushMsg(int16(arg.Ver), arg.Operation, arg.Msg); err != nil {
				return
			}
			reply.Index = int32(n)
		}
	}
	return
}

// Push push a message to a specified sub key, must goroutine safe.
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
		if channel = bucket.Get(key); channel != nil {
			if err = channel.PushMsg(int16(arg.Vers[n]), arg.Operations[n], arg.Msgs[n]); err != nil {
				return
			}
			reply.Index = int32(n)
		}
	}
	return
}

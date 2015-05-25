package main

import (
	log "code.google.com/p/log4go"
	"net"
	"net/rpc"
)

type RPCPushMsg struct {
	Key string
	Msg []byte
}

type RPCPushsMsg struct {
	Key        string
	Operations []int32
	Msgs       [][]byte
}

func InitRPCPush() error {
	c := &PushRPC{}
	rpc.Register(c)
	for _, bind := range Conf.RPCPushBind {
		log.Info("start listen rpc addr: \"%s\"", bind)
		go rpcListen(bind)
	}
	return nil
}

func rpcListen(bind string) {
	l, err := net.Listen("tcp", bind)
	if err != nil {
		log.Error("net.Listen(\"tcp\", \"%s\") error(%v)", bind, err)
		panic(err)
	}
	// if process exit, then close the rpc bind
	defer func() {
		log.Info("rpc addr: \"%s\" close", bind)
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
func (this *PushRPC) Push(args *RPCPushMsg, ret *int) (err error) {
	if args == nil {
		*ret = InternalErr
		log.Error("PushRPC.Push() args==nil")
		return
	}
	bucket := DefaultServer.Bucket(args.Key)
	if channel := bucket.Get(args.Key); channel != nil {
		// padding let caller do
		if err = channel.Push(1, OP_SEND_SMS_REPLY, args.Msg); err != nil {
			*ret = InternalErr
			log.Error("channel.Push() error(%v)", err)
			return
		}
	}
	*ret = OK
	return
}

// Pushs push multiple messages to a specified sub key, must goroutine safe.
func (this *PushRPC) Pushs(args *RPCPushsMsg, ret *int) (err error) {
	if args == nil {
		*ret = InternalErr
		log.Error("PushRPC.Pushs() args==nil")
		return
	}
	bucket := DefaultServer.Bucket(args.Key)
	if channel := bucket.Get(args.Key); channel != nil {
		// padding let caller do
		if _, err = channel.Pushs(1, args.Operations, args.Msgs); err != nil {
			*ret = InternalErr
			log.Error("channel.Pushs() error(%v)", err)
			return
		}
	}
	*ret = OK
	return
}

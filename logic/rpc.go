package main

import (
	log "code.google.com/p/log4go"
	inet "github.com/Terry-Mao/goim/libs/net"
	lproto "github.com/Terry-Mao/goim/proto/logic"
	rproto "github.com/Terry-Mao/goim/proto/router"
	rpc "github.com/Terry-Mao/protorpc"
	"net"
)

func InitRPC(auther Auther) (err error) {
	var (
		network, addr string
		c             = &RPC{auther: auther}
	)
	rpc.Register(c)
	for i := 0; i < len(Conf.RPCAddrs); i++ {
		log.Info("start listen rpc addr: \"%s\"", Conf.RPCAddrs[i])
		if network, addr, err = inet.ParseNetwork(Conf.RPCAddrs[i]); err != nil {
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
	// if process exit, then close the rpc bind
	defer func() {
		log.Info("rpc addr: \"%s\" close", addr)
		if err := l.Close(); err != nil {
			log.Error("listener.Close() error(%v)", err)
		}
	}()
	rpc.Accept(l)
}

// RPC
type RPC struct {
	auther Auther
}

// Connect auth and registe login
func (r *RPC) Connect(args *lproto.ConnArg, rep *lproto.ConnReply) (err error) {
	if args == nil {
		err = ErrArgs
		log.Error("Connect() error(%v)", err)
		return
	}
	// get userID from third implementation.
	// developer could implement "ThirdAuth" interface for decide how get userID
	userID := r.auther.Auth(args.Token)
	// notice router which connected
	c := RouterClient(userID)
	arg := &rproto.ConnArg{UserId: userID, Server: args.Server}
	reply := &rproto.ConnReply{}
	if err = c.Call(routerServiceConnect, arg, reply); err != nil {
		log.Error("c.Call(\"%s\",\"%v\") error(%s)", routerServiceConnect, *arg, err)
		return
	}
	rep.Key = Encode(userID, reply.Seq)
	return
}

// Disconnect notice router offline
func (r *RPC) Disconnect(args *lproto.DisconnArg, rep *lproto.DisconnReply) (err error) {
	if args == nil {
		err = ErrArgs
		log.Error("Disconnect() error(%v)", err)
		return
	}
	userID, seq, err := Decode(args.Key)
	if err != nil {
		log.Error("decode(\"%s\") error(%s)", args.Key, err)
		return
	}
	// notice router which disconnected
	c := RouterClient(userID)
	arg := &rproto.DisconnArg{UserId: userID, Seq: seq}
	reply := &rproto.DisconnReply{}
	if err = c.Call(routerServiceDisconnect, arg, reply); err != nil {
		log.Error("c.Call(\"%s\",\"%v\") error(%s)", routerServiceDisconnect, *arg, err)
		return
	}
	rep.Has = reply.Has
	return
}

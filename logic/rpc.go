package main

import (
	log "code.google.com/p/log4go"
	inet "github.com/Terry-Mao/goim/libs/net"
	lproto "github.com/Terry-Mao/goim/proto/logic"
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
	var (
		uid = r.auther.Auth(args.Token)
		seq int32
	)
	if seq, err = connect(uid, args.Server); err == nil {
		rep.Key = Encode(uid, seq)
	}
	return
}

// Disconnect notice router offline
func (r *RPC) Disconnect(args *lproto.DisconnArg, rep *lproto.DisconnReply) (err error) {
	if args == nil {
		err = ErrArgs
		log.Error("Disconnect() error(%v)", err)
		return
	}
	var (
		uid int64
		seq int32
	)
	if uid, seq, err = Decode(args.Key); err != nil {
		log.Error("decode(\"%s\") error(%s)", args.Key, err)
		return
	}
	rep.Has, err = disconnect(uid, seq)
	return
}

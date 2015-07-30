package main

import (
	log "code.google.com/p/log4go"
	"fmt"
	lproto "github.com/Terry-Mao/goim/proto/logic"
	rproto "github.com/Terry-Mao/goim/proto/router"
	rpc "github.com/Terry-Mao/protorpc"
	"net"
	"strconv"
	"strings"
)

func InitRPC(auther Auther) error {
	c := &RPC{auther: auther}
	rpc.Register(c)
	for _, bind := range Conf.RpcBind {
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

// RPC
type RPC struct {
	auther Auther
}

func (r *RPC) encode(userId int64, seq int32) string {
	return fmt.Sprintf("%d_%d", userId, seq)
}

func (r *RPC) decode(key string) (userId int64, seq int32, err error) {
	var (
		idx int
		t   int64
	)
	if idx = strings.IndexByte(key, '_'); idx == -1 {
		err = ErrDecodeKey
		return
	}
	if userId, err = strconv.ParseInt(key[:idx], 10, 64); err != nil {
		return
	}
	if t, err = strconv.ParseInt(key[idx+1:], 10, 32); err != nil {
		return
	}
	seq = int32(t)
	return
}

func (r *RPC) Ping(arg *lproto.PingArg, reply *lproto.PingReply) error {
	log.Debug("receive ping")
	return nil
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
	c := getRouterClient(userID)
	arg := &rproto.ConnArg{UserId: userID, Server: args.Server}
	reply := &rproto.ConnReply{}
	if err = c.Call(routerServiceConnect, arg, reply); err != nil {
		log.Error("c.Call(\"%s\",\"%v\") error(%s)", routerServiceConnect, *arg, err)
		return
	}
	rep.Key = r.encode(userID, reply.Seq)
	return
}

// Disconnect notice router offline
func (r *RPC) Disconnect(args *lproto.DisconnArg, rep *lproto.DisconnReply) (err error) {
	if args == nil {
		err = ErrArgs
		log.Error("Disconnect() error(%v)", err)
		return
	}
	userID, seq, err := r.decode(args.Key)
	if err != nil {
		log.Error("decode(\"%s\") error(%s)", args.Key, err)
		return
	}
	// notice router which disconnected
	c := getRouterClient(userID)
	arg := &rproto.DisconnArg{UserId: userID, Seq: seq}
	reply := &rproto.DisconnReply{}
	if err = c.Call(routerServiceDisconnect, arg, reply); err != nil {
		log.Error("c.Call(\"%s\",\"%v\") error(%s)", routerServiceDisconnect, *arg, err)
		return
	}
	rep.Has = reply.Has
	return
}

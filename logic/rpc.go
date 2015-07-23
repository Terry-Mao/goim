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

func InitRPC(thirdAuth ThirdAuth) error {
	c := &RPC{thirdAuth: thirdAuth}
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
	thirdAuth ThirdAuth
}

func (this *RPC) encode(userId int64, seq int32) string {
	return fmt.Sprintf("%d_%d", userId, seq)
}

func (this *RPC) decode(key string) (userId int64, seq int32, err error) {
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
	if t, err = strconv.ParseInt(key[idx:], 10, 32); err != nil {
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
func (this *RPC) Connect(args *lproto.ConnArg, rep *lproto.ConnReply) (err error) {
	if args == nil {
		err = fmt.Errorf("RPC.Connect() args==nil")
		log.Error(err)
		return
	}

	// get userID from third implementation.
	// developer could implement "ThirdAuth" interface for decide how get userID
	userID := this.thirdAuth.CheckUID(args.Token)

	// notice router which connected
	c := getRouterClient(userID)
	arg := &rproto.ConnArg{UserId: userID, Server: args.Server}
	reply := &rproto.ConnReply{}
	if err = c.Call(routerServiceConnect, arg, reply); err != nil {
		log.Error("c.Call(\"%s\",\"%v\") error(%s)", routerServiceConnect, *arg, err)
		return
	}
	rep.Key = this.encode(userID, reply.Seq)
	return
}

// Disconnect notice router offline
func (this *RPC) Disconnect(args *lproto.DisconnArg, rep *lproto.DisconnReply) (err error) {
	if args == nil {
		err = fmt.Errorf("RPC.Connect() args==nil")
		log.Error(err)
		return
	}
	userID, seq, err := this.decode(args.Key)
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

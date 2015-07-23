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
	arg := &rproto.ConnArg{UserId: userID, Server: args.Serverid}
	reply := &rproto.ConnReply{}
	if err = c.Call(routerServiceConnect, arg, reply); err != nil {
		log.Error("c.Call(\"%s\",\"%v\") error(%s)", routerServiceConnect, *arg, err)
		return
	}
	rep.Subkey = subKey(userID, reply.Seq)
	return
}

// subKey marshal subkey
func subKey(userID int64, seq int32) string {
	return fmt.Sprintf("%d_%d", userID, seq)
}

// unSubKey parse subkey
func unSubKey(subKeyStr string) (userID int64, seq int32, err error) {
	tmp := strings.Split(subKeyStr, "_")
	if len(tmp) != 2 {
		err = fmt.Errorf("subkey format error")
		return
	}
	userID, err = strconv.ParseInt(tmp[0], 10, 64)
	if err != nil {
		return
	}
	seq64, err := strconv.ParseInt(tmp[1], 10, 64)
	if err != nil {
		return
	}
	seq = int32(seq64)
	return
}

// Disconnect notice router offline
func (this *RPC) Disconnect(args *lproto.DisconnArg, rep *lproto.DisconnReply) (err error) {
	if args == nil {
		err = fmt.Errorf("RPC.Connect() args==nil")
		log.Error(err)
		return
	}
	userID, seq, err := unSubKey(args.Subkey)
	if err != nil {
		log.Error("unSubKey(\"%s\") error(%s)", args.Subkey, err)
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

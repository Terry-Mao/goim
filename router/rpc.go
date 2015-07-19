package main

import (
	log "code.google.com/p/log4go"
	"github.com/Terry-Mao/goim/router/proto"
	rpc "github.com/Terry-Mao/protorpc"
	"net"
)

func InitRPC(bs []*Bucket) error {
	c := &RouterRPC{Buckets: bs, BucketIdx: int64(len(bs)) - 1}
	rpc.Register(c)
	for _, bind := range Conf.RPCBind {
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

// Router RPC
type RouterRPC struct {
	Buckets   []*Bucket
	BucketIdx int64
}

func (r *RouterRPC) bucket(userId int64) *Bucket {
	idx := int(userId % r.BucketIdx)
	log.Debug("\"%d\" hit channel bucket index: %d", userId, idx)
	return r.Buckets[idx]
}

func (r *RouterRPC) Connect(arg *proto.ConnArg, reply *proto.ConnReply) error {
	reply.Seq = r.bucket(arg.UserId).Put(arg.UserId, arg.Server)
	return nil
}

func (r *RouterRPC) Disconnect(arg *proto.DisconnArg, reply *proto.DisconnReply) error {
	reply.Has = r.bucket(arg.UserId).DelSession(arg.UserId, arg.Seq)
	return nil
}

func (r *RouterRPC) Get(arg *proto.GetArg, reply *proto.GetReply) error {
	seqs, servers := r.bucket(arg.UserId).Get(arg.UserId)
	reply.Seqs = seqs
	reply.Servers = servers
	return nil
}

func (r *RouterRPC) MGet(arg *proto.MGetArg, reply *proto.MGetReply) error {
	var (
		i       int
		userId  int64
		session *proto.GetReply
	)
	reply.Sessions = make([]*proto.GetReply, 0, len(arg.UserIds))
	for i = 0; i < len(arg.UserIds); i++ {
		userId = arg.UserIds[i]
		seqs, servers := r.bucket(userId).Get(userId)
		session = new(proto.GetReply)
		session.Seqs = seqs
		session.Servers = servers
		reply.Sessions[i] = session
	}
	return nil
}

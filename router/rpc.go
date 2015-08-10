package main

import (
	log "code.google.com/p/log4go"
	rpc "github.com/Terry-Mao/protorpc"
	proto "github.com/thinkboy/goim/proto/router"
	"net"
)

func InitRPC(bs []*Bucket) error {
	c := &RouterRPC{Buckets: bs, BucketIdx: int64(len(bs))}
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

func (r *RouterRPC) GetAll(arg *proto.NoArg, reply *proto.GetAllReply) error {
	var (
		session *proto.GetReply
		userIds []int64
		i, j    = int64(0), 0
	)
	for i = 0; i < r.BucketIdx; i++ {
		userIds = r.Buckets[i].AllUsers()
		for j = 0; j < len(userIds); j++ {
			session = new(proto.GetReply)
			session.Seqs, session.Servers = r.bucket(userIds[j]).Get(userIds[j])
			reply.Sessions = append(reply.Sessions, session)
			log.Debug("seqs:%v servers:%v", session.Seqs, session.Servers)
		}
		reply.UserIds = append(reply.UserIds, userIds...)
	}
	return nil
}

func (r *RouterRPC) MGet(arg *proto.MGetArg, reply *proto.MGetReply) error {
	var (
		i       int
		userId  int64
		session *proto.GetReply
	)
	reply.Sessions = make([]*proto.GetReply, len(arg.UserIds))
	reply.UserIds = make([]int64, len(arg.UserIds))
	for i = 0; i < len(arg.UserIds); i++ {
		userId = arg.UserIds[i]
		session = new(proto.GetReply)
		session.Seqs, session.Servers = r.bucket(userId).Get(userId)
		reply.UserIds[i] = userId
		reply.Sessions[i] = session
		log.Debug("seqs:%v servers:%v", session.Seqs, session.Servers)
	}
	return nil
}

func (r *RouterRPC) GetSeqCount(arg *proto.GetSeqCountArg, reply *proto.GetSeqCountReply) error {
	reply.Count = int32(r.bucket(arg.UserId).Count(arg.UserId))
	return nil
}

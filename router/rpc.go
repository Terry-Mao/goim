package main

import (
	inet "goim/libs/net"
	"goim/libs/proto"
	"net"
	"net/rpc"

	log "github.com/thinkboy/log4go"
)

func InitRPC(bs []*Bucket) (err error) {
	var (
		network, addr string
		c             = &RouterRPC{Buckets: bs, BucketIdx: int64(len(bs))}
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

// Router RPC
type RouterRPC struct {
	Buckets   []*Bucket
	BucketIdx int64
}

func (r *RouterRPC) bucket(userId int64) *Bucket {
	idx := int(userId % r.BucketIdx)
	// fix panic
	if idx < 0 {
		idx = 0
	}
	return r.Buckets[idx]
}

func (r *RouterRPC) Ping(arg *proto.NoArg, reply *proto.NoReply) error {
	return nil
}

func (r *RouterRPC) Put(arg *proto.PutArg, reply *proto.PutReply) error {
	reply.Seq = r.bucket(arg.UserId).Put(arg.UserId, arg.Server, arg.RoomId)
	return nil
}

func (r *RouterRPC) Del(arg *proto.DelArg, reply *proto.DelReply) error {
	reply.Has = r.bucket(arg.UserId).Del(arg.UserId, arg.Seq, arg.RoomId)
	return nil
}

func (r *RouterRPC) DelServer(arg *proto.DelServerArg, reply *proto.NoReply) error {
	var (
		bucket *Bucket
	)
	for _, bucket = range r.Buckets {
		bucket.DelServer(arg.Server)
	}
	return nil
}

func (r *RouterRPC) Get(arg *proto.GetArg, reply *proto.GetReply) error {
	reply.Seqs, reply.Servers = r.bucket(arg.UserId).Get(arg.UserId)
	return nil
}

func (r *RouterRPC) GetAll(arg *proto.NoArg, reply *proto.GetAllReply) error {
	var (
		i             int64
		j             int
		userIds       []int64
		seqs, servers [][]int32
		session       *proto.GetReply
	)
	for i = 0; i < r.BucketIdx; i++ {
		userIds, seqs, servers = r.Buckets[i].GetAll()
		reply.UserIds = append(reply.UserIds, userIds...)
		for j = 0; j < len(userIds); j++ {
			session = new(proto.GetReply)
			session.Seqs, session.Servers = seqs[j], servers[j]
			reply.Sessions = append(reply.Sessions, session)
		}
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
	}
	return nil
}

func (r *RouterRPC) Count(arg *proto.NoArg, reply *proto.CountReply) error {
	var (
		bucket *Bucket
	)
	for _, bucket = range r.Buckets {
		reply.Count += bucket.Count()
	}
	return nil
}

func (r *RouterRPC) RoomCount(arg *proto.RoomCountArg, reply *proto.RoomCountReply) error {
	var (
		bucket *Bucket
	)
	for _, bucket = range r.Buckets {
		reply.Count += bucket.RoomCount(arg.RoomId)
	}
	return nil
}

func (r *RouterRPC) AllRoomCount(arg *proto.NoArg, reply *proto.AllRoomCountReply) error {
	var (
		bucket        *Bucket
		roomId, count int32
	)
	reply.Counter = make(map[int32]int32)
	for _, bucket = range r.Buckets {
		for roomId, count = range bucket.AllRoomCount() {
			reply.Counter[roomId] += count
		}
	}
	return nil
}

func (r *RouterRPC) AllServerCount(arg *proto.NoArg, reply *proto.AllServerCountReply) error {
	var (
		bucket        *Bucket
		server, count int32
	)
	reply.Counter = make(map[int32]int32)
	for _, bucket = range r.Buckets {
		for server, count = range bucket.AllServerCount() {
			reply.Counter[server] += count
		}
	}
	return nil
}

func (r *RouterRPC) UserCount(arg *proto.UserCountArg, reply *proto.UserCountReply) error {
	reply.Count = int32(r.bucket(arg.UserId).UserCount(arg.UserId))
	return nil
}

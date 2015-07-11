package main

import (
	log "code.google.com/p/log4go"
	rpc "github.com/Terry-Mao/goim/protorpc"
	"github.com/Terry-Mao/goim/router/proto"
	"net"
)

const (
	OK          = 1
	NoExistKey  = 65531
	ParamterErr = 65532
	InternalErr = 65535
)

func InitRPC() error {
	c := &RouterRPC{}
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
}

// Sub let client get sub info by sub key.
func (this *RouterRPC) Sub(key *proto.ArgKey, ret *proto.Ret) (err error) {
	if key == nil {
		log.Error("RouterRPC.Sub() key==nil")
		ret.Ret = ParamterErr
		return
	}
	sb := DefaultBuckets.SubBucket(key.Key)
	if sb == nil {
		log.Error("DefaultBuckets get subbucket error key(%s)", key.Key)
		ret.Ret = InternalErr
		return
	}
	n := sb.Get(key.Key)
	if n == nil {
		ret.Ret = NoExistKey
		return
	}
	ret.Ret |= (int64(n.server) << 48)
	ret.Ret |= (int64(n.state) << 32)
	ret.Ret |= OK
	return
}

// BatchSub let client batch get sub info by sub keys.
func (this *RouterRPC) BatchSub(key *proto.ArgBatchKey, ret *proto.RetBatchSub) (err error) {
	ret = new(proto.RetBatchSub)
	if key == nil {
		log.Error("RouterRPC.Push() key==nil")
		ret.Ret = ParamterErr
		return
	}
	l := len(key.Keys)
	if l == 0 {
		ret.Ret = OK
		return
	}
	ret.Subs = make([]*proto.ArgSub, l)
	i := 0
	for _, v := range key.Keys {
		sb := DefaultBuckets.SubBucket(v)
		if sb == nil {
			log.Error("DefaultBuckets batch get subbucket error key(%s)", v)
			continue
		}
		n := sb.Get(v)
		if n == nil {
			log.Error("DefaultBuckets batch get subbucket nil error key(%s)", v)
			continue
		}
		sub := &proto.ArgSub{}
		sub.Key = v
		sub.State = int32(n.state)
		sub.Server = int32(n.server)
		ret.Subs[i] = sub
		i++
	}
	ret.Subs = ret.Subs[:i]
	ret.Ret = OK
	return
}

// Topic let client get all sub key in topic.
func (this *RouterRPC) Topic(key *proto.ArgTopic, ret *proto.RetBatchSub) (err error) {
	ret = new(proto.RetBatchSub)
	if key == nil {
		log.Error("RouterRPC.Topic() key==nil")
		ret.Ret = ParamterErr
		return
	}
	tb := DefaultBuckets.TopicBucket(key.Topic)
	if tb == nil {
		log.Error("DefaultBuckets get topicbucket error key(%s)", key)
		ret.Ret = InternalErr
		return
	}
	m := tb.Get(key.Topic)
	l := len(m)
	if l > 0 {
		ret.Subs = make([]*proto.ArgSub, l)
		i := 0
		for k, _ := range m {
			ts := &proto.ArgSub{}
			ts.Key = k
			sb := DefaultBuckets.SubBucket(k)
			if sb == nil {
				continue
			}
			n := sb.Get(k)
			if n == nil {
				// TODO is or not delete from topics
				tb.del(key.Topic, k)
				continue
			}
			ts.State = int32(n.state)
			ts.Server = int32(n.server)
			ret.Subs[i] = ts
			i++
		}
		ret.Subs = ret.Subs[:i]
	}
	ret.Ret = OK
	return
}

// PbSetSub let client set sub key.
func (this *RouterRPC) SetSub(key *proto.ArgSub, ret *proto.Ret) (err error) {
	if key == nil {
		log.Error("RouterRPC.SetSub() key==nil")
		ret.Ret = ParamterErr
		return
	}
	DefaultBuckets.SubBucket(key.Key).SetStateAndServer(key.Key, int8(key.State), int16(key.Server))
	ret.Ret = OK
	return
}

// SetTopic let client set topic.
func (this *RouterRPC) SetTopic(key *proto.ArgTopicKey, ret *proto.Ret) (err error) {
	if key == nil {
		log.Error("RouterRPC.SetTopic() key==nil")
		ret.Ret = ParamterErr
		return
	}
	DefaultBuckets.PutToTopic(key.Topic, key.Key)
	ret.Ret = OK
	return
}

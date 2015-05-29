package main

import (
	log "code.google.com/p/log4go"
	pbcodec "github.com/felixhao/goim/router/protobuf"
	"net"
	"net/rpc"
	"time"
)

var (
	DEAD_LINE = 5 * time.Minute
)

type RPCSubMsg struct {
	Ret    int
	State  int8
	Server int16
}

type RPCBatchSubMsg struct {
	Ret  int
	Subs []*Sub
}

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
	// rpc.Accept(l)
	for {
		conn, err := l.Accept()
		if err != nil {
			continue
		}
		conn.SetDeadline(time.Now().Add(DEAD_LINE))
		go rpc.ServeCodec(pbcodec.NewPbServerCodec(conn))
	}
}

// Router RPC
type RouterRPC struct {
}

// Sub let client get sub info by sub key.
func (this *RouterRPC) Sub(key *string, ret *int64) (err error) {
	if key == nil {
		log.Error("RouterRPC.Push() key==nil")
		*ret = ParamterErr
		return
	}
	sb := DefaultBuckets.SubBucket(*key)
	if sb == nil {
		log.Error("DefaultBuckets get subbucket error key(%s)", *key)
		*ret = InternalErr
		return
	}
	n := sb.Get(*key)
	if n == nil {
		*ret = NoExistKey
		return
	}
	*ret |= (int64(n.server) << 48)
	*ret |= (int64(n.state) << 32)
	*ret |= OK
	return
}

// PbSub let client get sub info by sub key.
func (this *RouterRPC) PbSub(key *PbRPCSubKey, ret *PbRPCSubRet) (err error) {
	if key == nil {
		log.Error("RouterRPC.Push() key==nil")
		ret.Ret = ParamterErr
		return
	}
	log.Info("PbSub key(%s)", key.Key)
	sb := DefaultBuckets.SubBucket(key.Key)
	if sb == nil {
		log.Error("DefaultBuckets get subbucket error key(%s)", key.Key)
		ret.Ret = InternalErr
		return
	}
	n := sb.Get(key.Key)
	log.Info("PbSub node(%v)", n)
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
func (this *RouterRPC) BatchSub(key *[]string, ret *RPCBatchSubMsg) (err error) {
	ret = new(RPCBatchSubMsg)
	if key == nil {
		log.Error("RouterRPC.Push() key==nil")
		ret.Ret = ParamterErr
		return
	}
	l := len(*key)
	if l == 0 {
		ret.Ret = OK
		return
	}
	ret.Subs = make([]*Sub, l)
	i := 0
	for _, v := range *key {
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
		sub := &Sub{}
		sub.Key = v
		sub.State = n.state
		sub.Server = n.server
		ret.Subs[i] = sub
		i++
	}
	ret.Subs = ret.Subs[:i]
	ret.Ret = OK
	return
}

// Topic let client get all sub key in topic.
func (this *RouterRPC) Topic(key *string, ret *RPCBatchSubMsg) (err error) {
	ret = new(RPCBatchSubMsg)
	if key == nil {
		log.Error("RouterRPC.Topic() key==nil")
		ret.Ret = ParamterErr
		return
	}
	tb := DefaultBuckets.TopicBucket(*key)
	if tb == nil {
		log.Error("DefaultBuckets get topicbucket error key(%s)", *key)
		ret.Ret = InternalErr
		return
	}
	ret.Subs = tb.Get(*key)
	ret.Ret = OK
	return
}

type RPCTopicSubArg struct {
	Topic  string
	Subkey string
}

// SetTopic let client set topic.
func (this *RouterRPC) SetTopic(key *RPCTopicSubArg, ret *int) (err error) {
	if key == nil {
		log.Error("RouterRPC.SetTopic() key==nil")
		*ret = ParamterErr
		return
	}
	DefaultBuckets.PutToTopic(key.Topic, key.Subkey)
	*ret = OK
	return
}

type RPCSubArg struct {
	Subkey string
	State  int8
	Server int16
}

// SetSub let client set sub key.
func (this *RouterRPC) SetSub(key *RPCSubArg, ret *int) (err error) {
	if key == nil {
		log.Error("RouterRPC.SetSub() key==nil")
		*ret = ParamterErr
		return
	}
	log.Info("SetSub key(%s)", key)
	DefaultBuckets.SubBucket(key.Subkey).SetStateAndServer(key.Subkey, key.State, key.Server)
	*ret = OK
	return
}

// PbSetSub let client set sub key.
func (this *RouterRPC) PbSetSub(key *PbRPCSetSubArg, ret *PbRPCSubRet) (err error) {
	if key == nil {
		log.Error("RouterRPC.PbSetSub() key==nil")
		ret.Ret = ParamterErr
		return
	}
	log.Info("PbSetSub key(%v)", key)
	DefaultBuckets.SubBucket(key.Subkey).SetStateAndServer(key.Subkey, int8(key.State), int16(key.Server))
	ret.Ret = OK
	return
}

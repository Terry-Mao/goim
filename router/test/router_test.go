package test

import (
	proto "github.com/Terry-Mao/goim/proto/router"
	rpc "github.com/Terry-Mao/protorpc"
	"testing"
)

func TestRouterConnect(t *testing.T) {
	c, err := rpc.Dial("tcp", "localhost:7270")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	arg := &proto.ConnArg{}
	reply := &proto.ConnReply{}
	arg.UserId = 1
	arg.Server = 0
	if err = c.Call("RouterRPC.Connect", arg, reply); err != nil {
		t.Error(err)
		t.FailNow()
	}
	if reply.Seq != 1 {
		t.Errorf("reply seq: %d not equal 0", reply.Seq)
		t.FailNow()
	}
	arg.UserId = 1
	arg.Server = 0
	if err = c.Call("RouterRPC.Connect", arg, reply); err != nil {
		t.Error(err)
		t.FailNow()
	}
	if reply.Seq != 2 {
		t.Errorf("reply seq: %d not equal 1", reply.Seq)
		t.FailNow()
	}
}

func TestRouterDisconnect(t *testing.T) {
	c, err := rpc.Dial("tcp", "localhost:7270")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	arg := &proto.ConnArg{}
	reply := &proto.ConnReply{}
	arg.UserId = 2
	arg.Server = 0
	if err = c.Call("RouterRPC.Connect", arg, reply); err != nil {
		t.Error(err)
		t.FailNow()
	}
	if reply.Seq != 1 {
		t.Errorf("reply seq: %d not equal 0", reply.Seq)
		t.FailNow()
	}
	arg1 := &proto.DisconnArg{}
	arg1.UserId = 2
	arg1.Seq = 1
	reply1 := &proto.DisconnReply{}
	if err = c.Call("RouterRPC.Disconnect", arg1, reply1); err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !reply1.Has {
		t.Errorf("reply has: %d not equal true", reply1.Has)
		t.FailNow()
	}
}

func TestRouterGet(t *testing.T) {
	c, err := rpc.Dial("tcp", "localhost:7270")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	arg := &proto.GetArg{}
	reply := &proto.GetReply{}
	arg.UserId = 1
	if err = c.Call("RouterRPC.Get", arg, reply); err != nil {
		t.Error(err)
		t.FailNow()
	}
	if len(reply.Seqs) != 2 || len(reply.Servers) != 2 {
		t.Errorf("reply seqs||servers length not equals 2")
		t.FailNow()
	}
	if reply.Seqs[0] != 1 || reply.Seqs[1] != 2 {
		t.Error("reply seqs not match")
		t.FailNow()
	}
	if reply.Servers[0] != 0 || reply.Servers[1] != 0 {
		t.Errorf("reply servers not match, %v", reply.Servers)
		t.FailNow()
	}
}

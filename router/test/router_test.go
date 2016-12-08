package test

import (
	proto "goim/libs/proto"
	rpc "net/rpc"
	"sort"
	"testing"
)

func TestRouterPut(t *testing.T) {
	c, err := rpc.Dial("tcp", "localhost:7270")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	args := proto.PutArg{UserId: 1, Server: 0, RoomId: -1}
	reply := proto.PutReply{}
	if err = c.Call("RouterRPC.Put", &args, &reply); err != nil {
		t.Error(err)
		t.FailNow()
	}
	if reply.Seq != 1 {
		t.Errorf("reply seq: %d not equal 1", reply.Seq)
		t.FailNow()
	}
	if err = c.Call("RouterRPC.Put", &args, &reply); err != nil {
		t.Error(err)
		t.FailNow()
	}
	if reply.Seq != 2 {
		t.Errorf("reply seq: %d not equal 2", reply.Seq)
		t.FailNow()
	}
}

func TestRouterDel(t *testing.T) {
	c, err := rpc.Dial("tcp", "localhost:7270")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	args := proto.PutArg{UserId: 2, Server: 0, RoomId: -1}
	reply := proto.PutReply{}
	if err = c.Call("RouterRPC.Put", &args, &reply); err != nil {
		t.Error(err)
		t.FailNow()
	}
	if reply.Seq != 1 {
		t.Errorf("reply seq: %d not equal 1", reply.Seq)
		t.FailNow()
	}
	args1 := proto.DelArg{UserId: 2, Seq: 1, RoomId: -1}
	reply1 := proto.DelReply{}
	if err = c.Call("RouterRPC.Del", &args1, &reply1); err != nil {
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

	args := proto.GetArg{UserId: 1}
	reply := proto.GetReply{}
	if err = c.Call("RouterRPC.Get", &args, &reply); err != nil {
		t.Error(err)
		t.FailNow()
	}
	if len(reply.Seqs) != 2 || len(reply.Servers) != 2 {
		t.Errorf("reply seqs||servers length not equals 2")
		t.FailNow()
	}
	seqSize := len(reply.Seqs)
	seqs := make([]int, seqSize)
	for i := 0; i < seqSize; i++ {
		seqs[i] = int(reply.Seqs[i])
	}
	sort.Ints(seqs)
	if seqs[0] != 1 || seqs[1] != 2 {
		t.Error("reply seqs not match, %v", reply.Seqs)
		t.FailNow()
	}
	if reply.Servers[0] != 0 || reply.Servers[1] != 0 {
		t.Errorf("reply servers not match, %v", reply.Servers)
		t.FailNow()
	}
}

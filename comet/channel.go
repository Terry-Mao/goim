package main

import (
	"goim/libs/bufio"
	"goim/libs/define"
	"goim/libs/proto"
	"time"

	log "github.com/thinkboy/log4go"
)

var (
	ProtoReady  = &proto.Proto{Operation: define.OP_PROTO_READY}
	ProtoFinish = &proto.Proto{Operation: define.OP_PROTO_FINISH}
)

// Channel used by message pusher send msg to write goroutine.
type Channel struct {
	RoomId   int32
	CliProto Ring
	signal   chan *proto.Proto
	Writer   bufio.Writer
	Reader   bufio.Reader
}

func NewChannel(cli, svr int, rid int32) *Channel {
	c := new(Channel)
	c.RoomId = rid
	c.CliProto.Init(cli)
	c.signal = make(chan *proto.Proto, svr)
	return c
}

// Push server push message.
func (c *Channel) Push(p *proto.Proto) (err error) {
	select {
	case c.signal <- p:
	default:
		log.Error("lost message:%v roomid:%d", *p, c.RoomId)
	}
	return
}

// Ready check the channel ready or close?
func (c *Channel) Ready() (p *proto.Proto) {
	p = <-c.signal
	return
}

// ReadyWithTimeout check the channel ready or close?
func (c *Channel) ReadyWithTimeout(timeout time.Duration) (bool, *proto.Proto) {
	var p *proto.Proto
	select {
	case p = <-c.signal:
		return true, p
	case <-time.After(timeout):
		return false, nil
	}
	return false, nil
}

// Signal send signal to the channel, protocol ready.
func (c *Channel) Signal() {
	c.signal <- ProtoReady
}

// Close close the channel.
func (c *Channel) Close() {
	c.signal <- ProtoFinish
}

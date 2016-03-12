package main

import (
	"goim/libs/bufio"
	"goim/libs/proto"

	log "github.com/thinkboy/log4go"
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
		log.Error("lost a message:%s room:%d", p.Body, c.RoomId)
	}
	return
}

// Ready check the channel ready or close?
func (c *Channel) Ready() (p *proto.Proto) {
	p = <-c.signal
	return
}

// Signal send signal to the channel, protocol ready.
func (c *Channel) Signal() {
	c.signal <- proto.ProtoReady
}

// Close close the channel.
func (c *Channel) Close() {
	c.signal <- proto.ProtoFinish
}

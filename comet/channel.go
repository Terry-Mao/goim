package main

import (
	"sync"
	"time"
)

// Channel used by message pusher send msg to write goroutine.
type Channel struct {
	RoomId   int32
	signal   chan int
	CliProto Proto
	SvrProto Ring
	cLock    sync.Mutex
	Buf      [RawHeaderSize]byte
}

func NewChannel(proto int, rid int32) *Channel {
	c := new(Channel)
	c.RoomId = rid
	c.signal = make(chan int, SignalNum)
	c.SvrProto.Init(proto)
	return c
}

// Reply server reply message.
func (c *Channel) Reply(p *Proto) (err error) {
	var proto *Proto
	c.cLock.Lock()
	if proto, err = c.SvrProto.Set(); err == nil {
		*proto = *p
		c.SvrProto.SetAdv()
	}
	c.cLock.Unlock()
	c.Signal()
	return
}

// Push server push message.
func (c *Channel) Push(ver int16, operation int32, body []byte) (err error) {
	var proto *Proto
	c.cLock.Lock()
	if proto, err = c.SvrProto.Set(); err == nil {
		proto.Ver = ver
		proto.Operation = operation
		proto.Body = body
		c.SvrProto.SetAdv()
	}
	c.cLock.Unlock()
	c.Signal()
	return
}

// Pushs server push messages.
func (c *Channel) Pushs(ver []int32, operations []int32, bodies [][]byte) (idx int32, err error) {
	var (
		proto *Proto
		n     int32
	)
	c.cLock.Lock()
	for n = 0; n < int32(len(ver)); n++ {
		// fetch a proto from channel free list
		if proto, err = c.SvrProto.Set(); err == nil {
			proto.Ver = int16(ver[n])
			proto.Operation = operations[n]
			proto.Body = bodies[n]
			c.SvrProto.SetAdv()
			idx = n
		} else {
			break
		}
	}
	c.cLock.Unlock()
	c.Signal()
	return
}

// Ready check the channel ready or close?
func (c *Channel) Ready() bool {
	return (<-c.signal) == ProtoReady
}

// ReadyWithTimeout check the channel ready or close?
func (c *Channel) ReadyWithTimeout(timeout time.Duration) bool {
	var s int
	select {
	case s = <-c.signal:
		return s == ProtoReady
	case <-time.After(timeout):
		return false
	}
}

// Signal send signal to the channel, protocol ready.
func (c *Channel) Signal() {
	// just ignore duplication signal
	select {
	case c.signal <- ProtoReady:
	default:
	}
}

// Close close the channel.
func (c *Channel) Close() {
	select {
	case c.signal <- ProtoFinish:
	default:
	}
}

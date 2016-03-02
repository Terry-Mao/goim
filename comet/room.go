package main

import (
	"sync"

	itime "github.com/Terry-Mao/goim/libs/time"
	"github.com/thinkboy/goim/libs/proto"
)

type RoomOptions struct {
	ChannelSize int
}

type Room struct {
	id    int32
	rLock sync.RWMutex
	// map room id with channels
	// TODO use double-linked list
	chs map[*Channel]struct{}
}

// NewRoom new a room struct, store channel room info.
func NewRoom(id int32, t *itime.Timer, options RoomOptions) (r *Room) {
	r = new(Room)
	r.id = id
	r.chs = make(map[*Channel]struct{}, options.ChannelSize)
	return
}

// Put put channel into the room.
func (r *Room) Put(ch *Channel) {
	r.rLock.Lock()
	r.chs[ch] = struct{}{}
	r.rLock.Unlock()
	return
}

// Del delete channel from the room.
func (r *Room) Del(ch *Channel) {
	r.rLock.Lock()
	delete(r.chs, ch)
	r.rLock.Unlock()
}

// Push push msg to the room, if chan full discard it.
func (r *Room) Push(p *proto.Proto) {
	var ch *Channel
	r.rLock.RLock()
	for ch, _ = range r.chs {
		ch.Push(p)
	}
	r.rLock.RUnlock()
	return
}

// Online get online number.
func (r *Room) Online() (o int) {
	r.rLock.RLock()
	o = len(r.chs)
	r.rLock.RUnlock()
	return
}

// Close close the room.
func (r *Room) Close() {
	var ch *Channel
	r.rLock.RLock()
	for ch, _ = range r.chs {
		ch.Close()
	}
	r.rLock.RUnlock()
}

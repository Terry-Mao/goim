package main

import (
	"goim/libs/proto"
	"sync"
)

type Room struct {
	id     int32
	rLock  sync.RWMutex
	next   *Channel
	online int
	drop   bool
}

// NewRoom new a room struct, store channel room info.
func NewRoom(id int32) (r *Room) {
	r = new(Room)
	r.id = id
	r.drop = false
	r.next = nil
	r.online = 0
	return
}

// Put put channel into the room.
func (r *Room) Put(ch *Channel) (err error) {
	r.rLock.Lock()
	if !r.drop {
		if r.next != nil {
			r.next.Prev = ch
		}
		ch.Next = r.next
		ch.Prev = nil
		r.next = ch // insert to header
		r.online++
	} else {
		err = ErrRoomDroped
	}
	r.rLock.Unlock()
	return
}

// Del delete channel from the room.
func (r *Room) Del(ch *Channel) bool {
	r.rLock.Lock()
	// if header
	if r.next == ch {
		r.next = ch.Next
		if r.next != nil {
			r.next.Prev = nil // avoid memory leak
		}
	} else {
		if ch.Next != nil {
			// not footer node
			ch.Next.Prev = ch.Prev
		}
		ch.Prev.Next = ch.Next
	}
	r.online--
	r.drop = (r.online == 0)
	r.rLock.Unlock()
	return r.drop
}

// Push push msg to the room, if chan full discard it.
func (r *Room) Push(p *proto.Proto) {
	r.rLock.RLock()
	for ch := r.next; ch != nil; ch = ch.Next {
		ch.Push(p)
	}
	r.rLock.RUnlock()
	return
}

// Online get online number.
func (r *Room) Online() (o int) {
	r.rLock.RLock()
	o = r.online
	r.rLock.RUnlock()
	return
}

// Close close the room.
func (r *Room) Close() {
	r.rLock.RLock()
	for ch := r.next; ch != nil; ch = ch.Next {
		ch.Close()
	}
	r.rLock.RUnlock()
}

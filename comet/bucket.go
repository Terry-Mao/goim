package main

import (
	"goim/libs/define"
	"goim/libs/proto"
	"goim/libs/time"
	"sync"
	"sync/atomic"
)

type BucketOptions struct {
	ChannelSize   int
	RoomSize      int
	RoutineAmount int64
	RoutineSize   int
}

// Bucket is a channel holder.
type Bucket struct {
	cLock    sync.RWMutex        // protect the channels for chs
	chs      map[string]*Channel // map sub key to a channel
	boptions BucketOptions
	// room
	rooms       map[int32]*Room // bucket room channels
	routines    []chan *proto.BoardcastRoomArg
	routinesNum int64
	roptions    RoomOptions
}

// NewBucket new a bucket struct. store the key with im channel.
func NewBucket(boptions BucketOptions, roptions RoomOptions) (b *Bucket) {
	b = new(Bucket)
	b.chs = make(map[string]*Channel, boptions.ChannelSize)
	b.boptions = boptions

	//room
	b.rooms = make(map[int32]*Room, boptions.RoomSize)
	b.routines = make([]chan *proto.BoardcastRoomArg, boptions.RoutineAmount)
	b.routinesNum = int64(0)
	b.roptions = roptions
	for i := int64(0); i < boptions.RoutineAmount; i++ {
		c := make(chan *proto.BoardcastRoomArg, boptions.RoutineSize)
		b.routines[i] = c
		go b.roomproc(c)
	}
	return
}

// Put put a channel according with sub key.
func (b *Bucket) Put(key string, ch *Channel, tr *time.Timer) (err error) {
	var (
		room *Room
		ok   bool
	)
	b.cLock.Lock()
	b.chs[key] = ch
	if ch.RoomId != define.NoRoom {
		if room, ok = b.rooms[ch.RoomId]; !ok {
			room = NewRoom(ch.RoomId, tr, b.roptions)
			b.rooms[ch.RoomId] = room
		}
	}
	b.cLock.Unlock()
	if room != nil {
		err = room.Put(ch)
	}
	return
}

// Del delete the channel by sub key.
func (b *Bucket) Del(key string) {
	var (
		ok   bool
		drop bool
		ch   *Channel
		room *Room
	)
	b.cLock.Lock()
	if ch, ok = b.chs[key]; ok {
		delete(b.chs, key)
		if ch.RoomId != define.NoRoom {
			room, _ = b.rooms[ch.RoomId]
		}
	}
	b.cLock.Unlock()
	if room != nil {
		// if empty room, must delete from bucket
		if drop = room.Del(ch); drop {
			b.DelRoom(ch.RoomId)
		}
	}
}

// Channel get a channel by sub key.
func (b *Bucket) Channel(key string) (ch *Channel) {
	b.cLock.RLock()
	ch = b.chs[key]
	b.cLock.RUnlock()
	return
}

// Broadcast push msgs to all channels in the bucket.
func (b *Bucket) Broadcast(p *proto.Proto) {
	var ch *Channel
	b.cLock.RLock()
	for _, ch = range b.chs {
		// ignore error
		ch.Push(p)
	}
	b.cLock.RUnlock()
}

// Room get a room by roomid.
func (b *Bucket) Room(rid int32) (room *Room) {
	b.cLock.RLock()
	room, _ = b.rooms[rid]
	b.cLock.RUnlock()
	return
}

// DelRoom delete a room by roomid.
func (b *Bucket) DelRoom(rid int32) {
	var room *Room
	b.cLock.Lock()
	if room, _ = b.rooms[rid]; room != nil {
		delete(b.rooms, rid)
	}
	b.cLock.Unlock()
	if room != nil {
		room.Close()
	}
	return
}

// BroadcastRoom broadcast a message to specified room
func (b *Bucket) BroadcastRoom(arg *proto.BoardcastRoomArg) {
	num := atomic.AddInt64(&b.routinesNum, 1) % b.boptions.RoutineAmount
	b.routines[num] <- arg
}

// Rooms get all room id where online number > 0.
func (b *Bucket) Rooms() (res map[int32]struct{}) {
	var (
		roomId int32
		room   *Room
	)
	res = make(map[int32]struct{})
	b.cLock.RLock()
	for roomId, room = range b.rooms {
		if room.Online() > 0 {
			res[roomId] = struct{}{}
		}
	}
	b.cLock.RUnlock()
	return
}

// roomproc
func (b *Bucket) roomproc(c chan *proto.BoardcastRoomArg) {
	var (
		arg  *proto.BoardcastRoomArg
		room *Room
	)
	for {
		arg = <-c
		if room = b.Room(arg.RoomId); room != nil {
			room.Push(&arg.P)
		}
		arg = nil
		room = nil
	}
}

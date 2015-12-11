package main

import (
	"github.com/Terry-Mao/goim/libs/define"
	"sync"
)

type BucketOptions struct {
	ChannelSize int
	RoomSize    int
}

// Bucket is a channel holder.
type Bucket struct {
	cLock    sync.RWMutex        // protect the channels for chs
	chs      map[string]*Channel // map sub key to a channel
	rooms    map[int32]*Room     // bucket room channels
	boptions BucketOptions
	roptions RoomOptions
}

// NewBucket new a bucket struct. store the key with im channel.
func NewBucket(boptions BucketOptions, roptions RoomOptions) (b *Bucket) {
	b = new(Bucket)
	b.boptions = boptions
	b.roptions = roptions
	b.chs = make(map[string]*Channel, boptions.ChannelSize)
	b.rooms = make(map[int32]*Room, boptions.RoomSize)
	return
}

// Put put a channel according with sub key.
func (b *Bucket) Put(key string, ch *Channel) {
	var (
		room *Room
		ok   bool
	)
	b.cLock.Lock()
	b.chs[key] = ch
	if ch.RoomId != define.NoRoom {
		if room, ok = b.rooms[ch.RoomId]; !ok {
			room = NewRoom(ch.RoomId, b.roptions)
			b.rooms[ch.RoomId] = room
		}
	}
	b.cLock.Unlock()
	if room != nil {
		room.Put(ch)
	}
}

// Channel get a channel by sub key.
func (b *Bucket) Channel(key string) (ch *Channel) {
	b.cLock.RLock()
	ch = b.chs[key]
	b.cLock.RUnlock()
	return
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

// Del delete the channel by sub key.
func (b *Bucket) Del(key string) {
	var (
		ok   bool
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
		room.Del(ch)
		// TODO clean empty room
	}
}

// Broadcast push msgs to all channels in the bucket.
func (b *Bucket) Broadcast(ver int16, operation int32, msg []byte) {
	var ch *Channel
	b.cLock.RLock()
	for _, ch = range b.chs {
		// ignore error
		ch.PushMsg(ver, operation, msg)
	}
	b.cLock.RUnlock()
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

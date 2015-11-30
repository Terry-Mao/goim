package main

import (
	"github.com/Terry-Mao/goim/libs/define"
	"sync"
)

// Bucket is a channel holder.
type Bucket struct {
	cLock       sync.RWMutex                    // protect the channels for chs
	chs         map[string]*Channel             // map sub key to a channel
	rooms       map[int32]map[*Channel]struct{} // map room id with channels
	roomChannel int
}

// NewBucket new a bucket struct. store the subkey with im channel.
func NewBucket(channel, room, roomChannel, cliProto, svrProto int) *Bucket {
	b := new(Bucket)
	b.chs = make(map[string]*Channel, channel)
	b.rooms = make(map[int32]map[*Channel]struct{}, room)
	b.roomChannel = roomChannel
	return b
}

// Put put a channel according with sub key.
func (b *Bucket) Put(subKey string, ch *Channel) {
	var (
		room map[*Channel]struct{}
		ok   bool
	)
	b.cLock.Lock()
	b.chs[subKey] = ch
	if ch.RoomId != define.NoRoom {
		if room, ok = b.rooms[ch.RoomId]; !ok {
			room = make(map[*Channel]struct{}, b.roomChannel)
			b.rooms[ch.RoomId] = room
		}
		room[ch] = struct{}{}
	}
	b.cLock.Unlock()
}

// Get get a channel by sub key.
func (b *Bucket) Get(subKey string) *Channel {
	var ch *Channel
	b.cLock.RLock()
	ch = b.chs[subKey]
	b.cLock.RUnlock()
	return ch
}

// Del delete the channel by sub key.
func (b *Bucket) Del(subKey string) {
	var (
		ok   bool
		ch   *Channel
		room map[*Channel]struct{}
	)
	b.cLock.Lock()
	if ch, ok = b.chs[subKey]; ok {
		delete(b.chs, subKey)
		if ch.RoomId != define.NoRoom {
			if room, ok = b.rooms[ch.RoomId]; ok {
				// clean the room's channel
				// when room empty
				// clean the room space for free large room memory
				// WARN: if room flip between empty and someone let GC busy
				// this scene is rare
				delete(room, ch)
				if len(room) == 0 {
					delete(b.rooms, ch.RoomId)
				}
			}
		}
	}
	b.cLock.Unlock()
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

// Broadcast push msgs to all channels in the bucket's room.
func (b *Bucket) BroadcastRoom(rid int32, ver int16, operation int32, msg []byte) {
	var (
		ok   bool
		ch   *Channel
		room map[*Channel]struct{}
	)
	b.cLock.RLock()
	if room, ok = b.rooms[rid]; ok && len(room) > 0 {
		for ch, _ = range room {
			// ignore error
			ch.PushMsg(ver, operation, msg)
		}
	}
	b.cLock.RUnlock()
}

// Rooms get all room id where online number > 0.
func (b *Bucket) Rooms() (res map[int32]struct{}) {
	var (
		roomId int32
		room   map[*Channel]struct{}
	)
	b.cLock.RLock()
	res = make(map[int32]struct{}, len(b.rooms))
	for roomId, room = range b.rooms {
		if len(room) > 0 {
			res[roomId] = struct{}{}
		}
	}
	b.cLock.RUnlock()
	return
}

package main

import (
	"sync"
)

// Bucket is a channel holder.
type Bucket struct {
	cLock sync.RWMutex        // protect the channels for chs
	chs   map[string]*Channel // map sub key to a channel
}

// NewBucket new a bucket struct. store the subkey with im channel.
func NewBucket(channel, cliProto, svrProto int) *Bucket {
	b := new(Bucket)
	b.chs = make(map[string]*Channel, channel)
	return b
}

// Put put a channel according with sub key.
func (b *Bucket) Put(subKey string, ch *Channel) {
	b.cLock.Lock()
	b.chs[subKey] = ch
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
	b.cLock.Lock()
	delete(b.chs, subKey)
	b.cLock.Unlock()
}

// Broadcast push msgs to all channels in the bucket.
func (b *Bucket) Broadcast(ver int16, operation int32, msg []byte) {
	var (
		i   = 0
		ch  *Channel
		chs []*Channel
	)
	b.cLock.RLock()
	chs = make([]*Channel, len(b.chs))
	for _, ch = range b.chs {
		chs[i] = ch
		i++
	}
	b.cLock.RUnlock()
	for _, ch = range chs {
		// ignore error
		ch.PushMsg(ver, operation, msg)
	}
}

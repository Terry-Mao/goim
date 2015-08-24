package main

import (
	//	log "code.google.com/p/log4go"
	"sync"
)

const (
	bucketIncr = 10000
)

// Bucket is a channel holder.
type Bucket struct {
	cLock sync.Mutex          // protect the channels for chs
	chs   map[string]*Channel // map sub key to a channel
}

// NewBucket new a bucket struct. store the subkey with im channel.
func NewBucket(channel, cliProto, svrProto int) *Bucket {
	b := new(Bucket)
	b.chs = make(map[string]*Channel, channel)
	return b
}

// Put put a channel according with sub key and return the old one.
func (b *Bucket) Put(subKey string, ch *Channel) {
	b.cLock.Lock()
	b.chs[subKey] = ch
	b.cLock.Unlock()
}

// Get get a channel by sub key.
func (b *Bucket) Get(subKey string) *Channel {
	var ch *Channel
	b.cLock.Lock()
	ch = b.chs[subKey]
	b.cLock.Unlock()
	return ch
}

// Del delete the channel by sub key.
func (b *Bucket) Del(subKey string) {
	b.cLock.Lock()
	delete(b.chs, subKey)
	b.cLock.Unlock()
}

func (b *Bucket) Boardcast(ver int16, operation int32, msg []byte) {
	var (
		chl int
		ch  *Channel
		chs []*Channel
	)
	b.cLock.Lock()
	chl = len(b.chs)
	b.cLock.Unlock()
	// copy all channels
	// WARN: chl is dirty data, we add more channel slice avoid realloc slice
	chs = make([]*Channel, 0, chl+bucketIncr)
	b.cLock.Lock()
	for _, ch = range b.chs {
		chs = append(chs, ch)
	}
	b.cLock.Unlock()
	for _, ch = range chs {
		// ignore error
		ch.PushMsg(ver, operation, msg)
	}
}

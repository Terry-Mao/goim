package main

import (
	//	log "code.google.com/p/log4go"
	"sync"
)

// Bucket is a channel holder.
type Bucket struct {
	cLock sync.Mutex          // protect the free list of channel
	chs   map[string]*Channel // map sub key to a channel
	//bLock sync.Mutex          // protect the channels for chs
	//free  *Channel            // channel free list, reuse channel for everyone
	//used  int                 // count the used channl
}

// NewBucket new a bucket struct. store the subkey with im channel.
func NewBucket(channel, cliProto, svrProto int) *Bucket {
	b := new(Bucket)
	b.chs = make(map[string]*Channel, channel)
	//log.Debug("create %d bucket for store sub channel, each channel has cli.proto %d, svr.proto %d", channel, cliProto, svrProto)
	/*
		// pre alloc channel
		ch := NewChannel(cliProto, svrProto)
		b.free = ch
		for i := 1; i < channel; i++ {
			ch.next = NewChannel(cliProto, svrProto)
			ch = ch.next
		}
	*/
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

/*
// GetChannel get a empty channel use it's free list.
func (b *Bucket) GetChannel() *Channel {
	b.cLock.Lock()
	ch := b.free
	if ch != nil {
		b.free = ch.next
		*ch = Channel{} // reset
		b.used++
		log.Debug("get channel, used: %d", b.used)
	} else {
		log.Debug("bucket empty")
		ch = new(Channel)
	}
	b.cLock.Unlock()
	return ch
}

// PutChannel return back the ch to the free list.
func (b *Bucket) PutChannel(ch *Channel) {
	// if no used channel, free list full, discard it
	if b.used == 0 {
		log.Debug("bucket full")
		return
	}
	b.cLock.Lock()
	// double check
	if b.used == 0 {
		b.cLock.Unlock()
		log.Debug("bucket full")
		return
	}
	b.used--
	log.Debug("put channel, used: %d", b.used)
	ch.next = b.free
	b.free = ch
	b.cLock.Unlock()
}
*/

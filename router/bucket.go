package main

import (
	"sync"
	"time"
)

type Bucket struct {
	bLock    sync.RWMutex       // protect the session map
	sessions map[int64]*Session // map[user_id] ->  map[sub_id] -> server_id
	server   int
	cleaner  *Cleaner
}

// NewBucket new a bucket struct. store the subkey with im channel.
func NewBucket(session, server, cleaner int) *Bucket {
	b := new(Bucket)
	b.sessions = make(map[int64]*Session, session)
	b.server = server
	b.cleaner = NewCleaner(cleaner)
	go b.clean()
	return b
}

// Put put a channel according with user id.
func (b *Bucket) Put(userId int64, server int32) (seq int32) {
	var (
		s  *Session
		ok bool
	)
	b.bLock.Lock()
	if s, ok = b.sessions[userId]; !ok {
		s = NewSession(b.server)
		b.sessions[userId] = s
	}
	seq = s.Put(server)
	b.bLock.Unlock()
	return
}

func (b *Bucket) Get(userId int64) (seqs []int32, servers []int32) {
	var (
		s      *Session
		seq    int32
		server int32
		ok     bool
	)
	b.bLock.RLock()
	if s, ok = b.sessions[userId]; ok {
		seqs = make([]int32, 0, len(s.Servers()))
		servers = make([]int32, 0, len(s.Servers()))
		for seq, server = range s.Servers() {
			seqs = append(seqs, seq)
			servers = append(servers, server)
		}
	}
	b.bLock.RUnlock()
	return
}

func (b *Bucket) AllUsers() (userIds []int64) {
	b.bLock.RLock()
	userIds = make([]int64, 0, len(b.sessions))
	for userId, _ := range b.sessions {
		userIds = append(userIds, userId)
	}
	b.bLock.RUnlock()
	return
}

func (b *Bucket) Count(userId int64) (count int) {
	b.bLock.RLock()
	if s, ok := b.sessions[userId]; ok {
		count = s.Size()
	}
	b.bLock.RUnlock()
	return
}

func (b *Bucket) del(userId int64) {
	var (
		s  *Session
		ok bool
	)
	if s, ok = b.sessions[userId]; ok {
		if s.Size() == 0 {
			delete(b.sessions, userId)
		}
	}
}

func (b *Bucket) Del(userId int64) {
	b.bLock.Lock()
	b.del(userId)
	b.bLock.Unlock()
}

// Del delete the channel by sub key.
func (b *Bucket) DelSession(userId int64, seq int32) (ok bool) {
	var (
		s     *Session
		empty bool
	)
	b.bLock.Lock()
	if s, ok = b.sessions[userId]; ok {
		// WARN:
		// delete(b.sessions, userId)
		// empty is a dirty data, we use here for try lru clean discard session.
		// when one user flapped connect & disconnect, this also can reduce
		// frequently new & free object, gc is slow!!!
		empty = s.Del(seq)
	}
	b.bLock.Unlock()
	// lru
	if empty {
		b.cleaner.PushFront(userId, Conf.SessionExpire)
	}
	return
}

func (b *Bucket) clean() {
	var (
		i       int
		userIds []int64
	)
	for {
		userIds = b.cleaner.Clean()
		if len(userIds) != 0 {
			b.bLock.Lock()
			for i = 0; i < len(userIds); i++ {
				b.del(userIds[i])
			}
			b.bLock.Unlock()
			continue
		}
		time.Sleep(Conf.BucketCleanPeriod)
	}
}

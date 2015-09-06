package main

import (
	"github.com/Terry-Mao/goim/define"
	"sync"
	"time"
)

type Bucket struct {
	bLock    sync.RWMutex       // protect the session map
	sessions map[int64]*Session // map[user_id] ->  map[sub_id] -> server_id
	counter  map[int32]int32    // map[roomid] -> count, if noroom then all
	cleaner  *Cleaner
	server   int // session cache server number
}

// NewBucket new a bucket struct. store the subkey with im channel.
func NewBucket(session, server, cleaner int) *Bucket {
	b := new(Bucket)
	b.sessions = make(map[int64]*Session, session)
	b.counter = make(map[int32]int32)
	b.server = server
	b.cleaner = NewCleaner(cleaner)
	go b.clean()
	return b
}

// Put put a channel according with user id.
func (b *Bucket) Put(userId int64, server int32, roomId int32) (seq int32) {
	var (
		s  *Session
		ok bool
	)
	b.bLock.Lock()
	if s, ok = b.sessions[userId]; !ok {
		s = NewSession(b.server)
		b.sessions[userId] = s
	}
	if roomId != define.NoRoom {
		seq = s.PutRoom(server, roomId)
	} else {
		seq = s.Put(server)
	}
	b.counter[roomId]++
	b.bLock.Unlock()
	return
}

func (b *Bucket) Get(userId int64) (seqs []int32, servers []int32) {
	b.bLock.RLock()
	if s, ok := b.sessions[userId]; ok {
		seqs, servers = s.Servers()
	}
	b.bLock.RUnlock()
	return
}

func (b *Bucket) GetAll() (userIds []int64, seqs [][]int32, servers [][]int32) {
	b.bLock.RLock()
	i := len(b.sessions)
	userIds = make([]int64, i)
	seqs = make([][]int32, i)
	servers = make([][]int32, i)
	for userId, s := range b.sessions {
		i--
		userIds[i] = userId
		seqs[i], servers[i] = s.Servers()
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
		if s.Count() == 0 {
			delete(b.sessions, userId)
		}
	}
}

//func (b *Bucket) Del(userId int64) {
//	b.bLock.Lock()
//	b.del(userId)
//	b.bLock.Unlock()
//}

// Del delete the channel by sub key.
func (b *Bucket) Del(userId int64, seq int32, roomId int32) (ok bool) {
	var (
		s          *Session
		has, empty bool
	)
	b.bLock.Lock()
	if s, ok = b.sessions[userId]; ok {
		// WARN:
		// delete(b.sessions, userId)
		// empty is a dirty data, we use here for try lru clean discard session.
		// when one user flapped connect & disconnect, this also can reduce
		// frequently new & free object, gc is slow!!!
		if roomId != define.NoRoom {
			has, empty = s.DelRoom(seq, roomId)
		} else {
			has, empty = s.Del(seq)
		}
		if has {
			b.counter[roomId]--
		}
	}
	b.bLock.Unlock()
	// lru
	if empty {
		b.cleaner.PushFront(userId, Conf.SessionExpire)
	}
	return
}

func (b *Bucket) count(roomId int32) (count int32) {
	b.bLock.RLock()
	count = b.counter[roomId]
	b.bLock.RUnlock()
	return
}

func (b *Bucket) Count() (count int32) {
	count = b.count(define.NoRoom)
	return
}

func (b *Bucket) RoomCount(roomId int32) (count int32) {
	count = b.count(roomId)
	return
}

func (b *Bucket) AllRoomCount() (counter map[int32]int32) {
	var roomId, count int32
	b.bLock.RLock()
	counter = make(map[int32]int32, len(b.counter))
	for roomId, count = range b.counter {
		if roomId != define.NoRoom {
			counter[roomId] = count
		}
	}
	b.bLock.RUnlock()
	return
}

func (b *Bucket) UserCount(userId int64) (count int32) {
	b.bLock.RLock()
	if s, ok := b.sessions[userId]; ok {
		count = int32(s.Count())
	}
	b.bLock.RUnlock()
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

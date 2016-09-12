package main

import (
	"goim/libs/define"
	"sync"
	"time"
)

type Bucket struct {
	bLock             sync.RWMutex
	server            int                       // session server map init num
	session           int                       // bucket session init num
	sessions          map[int64]*Session        // userid->sessions
	roomCounter       map[int32]int32           // roomid->count
	serverCounter     map[int32]int32           // server->count
	userServerCounter map[int32]map[int64]int32 // serverid->userid count
	cleaner           *Cleaner                  // bucket map cleaner
}

// NewBucket new a bucket struct. store the subkey with im channel.
func NewBucket(session, server, cleaner int) *Bucket {
	b := new(Bucket)
	b.sessions = make(map[int64]*Session, session)
	b.roomCounter = make(map[int32]int32)
	b.serverCounter = make(map[int32]int32)
	b.userServerCounter = make(map[int32]map[int64]int32)
	b.cleaner = NewCleaner(cleaner)
	b.server = server
	b.session = session
	go b.clean()
	return b
}

// counter incr or decr counter.
func (b *Bucket) counter(userId int64, server int32, roomId int32, incr bool) {
	var (
		sm map[int64]int32
		v  int32
		ok bool
	)
	if sm, ok = b.userServerCounter[server]; !ok {
		sm = make(map[int64]int32, b.session)
		b.userServerCounter[server] = sm
	}
	if incr {
		sm[userId]++
		b.roomCounter[roomId]++
		b.serverCounter[server]++
	} else {
		// WARN:
		// if decr a userid but key not exists just ignore
		// this may not happen
		if v, _ = sm[userId]; v-1 == 0 {
			delete(sm, userId)
		} else {
			sm[userId] = v - 1
		}
		b.roomCounter[roomId]--
		b.serverCounter[server]--
	}
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
	b.counter(userId, server, roomId, true)
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

// Del delete the channel by sub key.
func (b *Bucket) Del(userId int64, seq int32, roomId int32) (ok bool) {
	var (
		s          *Session
		server     int32
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
			has, empty, server = s.DelRoom(seq, roomId)
		} else {
			has, empty, server = s.Del(seq)
		}
		if has {
			b.counter(userId, server, roomId, false)
		}
	}
	b.bLock.Unlock()
	// lru
	if empty {
		b.cleaner.PushFront(userId, Conf.SessionExpire)
	}
	return
}

func (b *Bucket) DelServer(server int32) {
	var (
		roomCounter       = make(map[int32]int32)
		servers           map[int32]int32
		userServerCounter map[int64]int32
		roomId            int32
		count             int32
		userId            int64
		s                 *Session
		ok                bool
	)
	b.bLock.Lock()
	// if server failed, may not accept new connections, just delete map
	if userServerCounter, ok = b.userServerCounter[server]; ok {
		delete(b.userServerCounter, server)
	} else {
		b.bLock.Unlock()
		return
	}
	delete(b.serverCounter, server)
	for userId, _ = range userServerCounter {
		if s, ok = b.sessions[userId]; !ok {
			continue
		}
		for roomId, servers = range s.rooms {
			roomCounter[roomId] += int32(len(servers))
		}
		delete(b.sessions, userId)
	}
	for roomId, count = range roomCounter {
		b.roomCounter[roomId] -= count
	}
	b.bLock.Unlock()
	return
}

func (b *Bucket) count(roomId int32) (count int32) {
	b.bLock.RLock()
	count = b.roomCounter[roomId]
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

func (b *Bucket) AllRoomCount() (roomCounter map[int32]int32) {
	var roomId, count int32
	b.bLock.RLock()
	roomCounter = make(map[int32]int32, len(b.roomCounter))
	for roomId, count = range b.roomCounter {
		if count > 0 {
			roomCounter[roomId] = count
		}
	}
	b.bLock.RUnlock()
	return
}

func (b *Bucket) AllServerCount() (serverCounter map[int32]int32) {
	var server, count int32
	b.bLock.RLock()
	serverCounter = make(map[int32]int32, len(b.serverCounter))
	for server, count = range b.serverCounter {
		serverCounter[server] = count
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

func (b *Bucket) delEmpty(userId int64) {
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
				b.delEmpty(userIds[i])
			}
			b.bLock.Unlock()
			continue
		}
		time.Sleep(Conf.BucketCleanPeriod)
	}
}

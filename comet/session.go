package main

import (
	"crypto/rand"
	"sync"
	"time"
)

var (
	SessionExpire = time.Hour * 24 * 7
	SessionDelay  = 1 * time.Second
)

const (
	sessionIdLen       = 16
	sessionBatchExpire = 100
)

// use hashmap + double linked list(LRU) + lazy drop
type SessionData struct {
	sid        string
	ki         []byte
	expireTime time.Time
	next, prev *SessionData
}

type SessionLRU struct {
	root SessionData
	len  int
}

type Session struct {
	lock     sync.Mutex
	sessions map[string]*SessionData
	lru      SessionLRU
}

func NewSession(sessionSize int) *Session {
	s := &Session{sessions: make(map[string]*SessionData, sessionSize)}
	s.lru.root.next = &s.lru.root
	s.lru.root.prev = &s.lru.root
	s.lru.len = 0
	return s
}

func (d *SessionData) Expire() bool {
	return d.expireTime.Before(time.Now())
}

func (l *SessionLRU) PushFront(e *SessionData) {
	at := &l.root
	n := at.next
	at.next = e
	e.prev = at
	e.next = n
	n.prev = e
	l.len++
}

func (l *SessionLRU) remove(e *SessionData) {
	e.prev.next = e.next
	e.next.prev = e.prev
	e.next = nil // avoid memory leaks
	e.prev = nil // avoid memory leaks
	l.len--
}

func (l *SessionLRU) MoveToFront(e *SessionData) {
	if l.root.next == e {
		return
	}
	at := &l.root
	// remove element
	e.prev.next = e.next
	e.next.prev = e.prev
	n := at.next
	at.next = e
	e.prev = at
	e.next = n
	n.prev = e
}

func (l *SessionLRU) Back() *SessionData {
	if l.len == 0 {
		return nil
	}
	return l.root.prev
}

func (s *Session) SessionId() string {
	var (
		err    error
		b1     byte
		b2     byte
		res    = make([]byte, sessionIdLen*2)
		random = res[sessionIdLen:]
	)
	if _, err = rand.Read(random); err != nil {
		return ""
	}
	for j := 0; j < sessionIdLen; j++ {
		b1 = (byte)((random[j] & 0xf0) >> 4)
		b2 = (byte)(random[j] & 0x0f)
		if b1 < 10 {
			res[j*2] = (byte)('0' + b1)
		} else {
			res[j*2] = (byte)('A' + (b1 - 10))
		}
		if b2 < 10 {
			res[j*2+1] = (byte)('0' + b2)
		} else {
			res[j*2+1] = (byte)('A' + (b2 - 10))
		}
	}
	return string(res)
}

func (s *Session) Put(ki []byte, expire time.Duration) (sid string) {
	var (
		ok bool
		sd = &SessionData{ki: ki}
	)
	s.lock.Lock()
	for {
		// while not exist
		// new session id
		if sid = s.SessionId(); sid == "" {
			continue
		}
		if _, ok = s.sessions[sid]; !ok {
			sd.expireTime = time.Now().Add(expire)
			sd.sid = sid
			s.sessions[sid] = sd
			// insert lru
			s.lru.PushFront(sd)
			break
		}
	}
	s.lock.Unlock()
	return
}

func (s *Session) Get(sid string) (ki []byte) {
	var (
		sd *SessionData
		ok bool
	)
	s.lock.Lock()
	if sd, ok = s.sessions[sid]; ok {
		if sd.Expire() {
			// lazy-drop
			delete(s.sessions, sid)
			s.lru.remove(sd)
		} else {
			// update lru
			s.lru.MoveToFront(sd)
			ki = sd.ki
		}
	}
	s.lock.Unlock()
	return
}

func (s *Session) clean() {
	var i int
	s.lock.Lock()
	for {
		e := s.lru.Back()
		if e == nil {
			s.lock.Unlock()
			return
		}
		if e.Expire() {
			s.lru.remove(e)
			delete(s.sessions, e.sid)
		}
		if i++; i >= sessionBatchExpire {
			// next time
			s.lock.Unlock()
			return
		}
	}
}

func SessionProcess(sessions []*Session) {
	for {
		for _, s := range sessions {
			s.clean()
		}
		time.Sleep(SessionDelay)
	}
}

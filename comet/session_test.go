package main

import (
	log "code.google.com/p/log4go"
	"testing"
	"time"
)

func TestSessionId(t *testing.T) {
	s := NewSession(10)
	log.Debug("sessionid: %s", s.SessionId())
}

func TestSession(t *testing.T) {
	key := "test"
	s := NewSession(10)
	log.Debug("lru: %x", s.lru)
	ki := s.Get(key)
	if ki != nil {
		t.FailNow()
	}
	sid := s.Put([]byte("test"), 1*time.Second)
	ki = s.Get(sid)
	if ki == nil || string(ki) != "test" {
		t.FailNow()
	}
}

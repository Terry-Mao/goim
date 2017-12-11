package main

import (
	"sync/atomic"
	"time"
)

type Stat struct {
	// msg
	MsgSucceeded uint64 `json:"msg_succeeded"`
	MsgFailed    uint64 `json:"msg_failed"`
	// sync
	SyncTimes uint64 `json:"sync_times"`
	// speed
	SpeedMsgSecond uint64 `json:"speed_msg_second"`
	// nodes
	RouterNodes map[string]string `json:"router_nodes"`
}

func NewStat() *Stat {
	s := new(Stat)
	go s.procSpeed()
	return s
}

func (s *Stat) Info() *Stat {
	s.RouterNodes = Conf.RouterRPCAddrs
	return s
}

func (s *Stat) Reset() {
	atomic.StoreUint64(&s.MsgSucceeded, 0)
	atomic.StoreUint64(&s.MsgFailed, 0)
	atomic.StoreUint64(&s.SyncTimes, 0)
}

func (s *Stat) procSpeed() {
	var (
		timer   = uint64(5) // diff 5s
		lastMsg uint64
	)
	for {
		s.SpeedMsgSecond = (atomic.LoadUint64(&s.MsgSucceeded) - lastMsg) / timer
		lastMsg = s.MsgSucceeded
		time.Sleep(time.Duration(timer) * time.Second)
	}
}

func (s *Stat) IncrMsgSucceeded() {
	atomic.AddUint64(&s.MsgSucceeded, 1)
}

func (s *Stat) IncrMsgFailed() {
	atomic.AddUint64(&s.MsgFailed, 1)
}

func (s *Stat) IncrSyncTimes() {
	atomic.AddUint64(&s.SyncTimes, 1)
}

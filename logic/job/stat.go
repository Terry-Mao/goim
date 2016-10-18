package main

import (
	"sync/atomic"
	"time"
)

type Stat struct {
	// messages
	AllMsg           uint64 `json:"all_msg"`
	PushMsg          uint64 `json:"push_msg"`
	BroadcastMsg     uint64 `json:"broadcast_msg"`
	BroadcastRoomMsg uint64 `json:"broadcast_room_msg"`
	// miss
	PushMsgFailed          uint64 `json:"push_msg_failed"`
	BroadcastMsgFailed     uint64 `json:"broadcast_msg_failed"`
	BroadcastRoomMsgFailed uint64 `json:"broadcast_room_msg_failed"`
	// speed
	SpeedMsgSecond       uint64 `json:"speed_msg_second"`
	SpeedRoomBatchSecond uint64 `json:"speed_room_batch_second"`
	// room
	ActiveRoomCount int `json:"active_room_count"`
	// nodes
	CometNodes map[int32]string `json:"comet_nodes"`
}

func NewStat() *Stat {
	s := new(Stat)
	go s.procSpeed()
	return s
}

func (s *Stat) Info() *Stat {
	s.ActiveRoomCount = roomBucket.Size()
	s.CometNodes = Conf.Comets
	return s
}

func (s *Stat) Reset() {
	atomic.StoreUint64(&s.AllMsg, 0)
	atomic.StoreUint64(&s.PushMsg, 0)
	atomic.StoreUint64(&s.BroadcastMsg, 0)
	atomic.StoreUint64(&s.BroadcastRoomMsg, 0)
	atomic.StoreUint64(&s.PushMsgFailed, 0)
	atomic.StoreUint64(&s.BroadcastMsgFailed, 0)
	atomic.StoreUint64(&s.BroadcastRoomMsgFailed, 0)
}

func (s *Stat) procSpeed() {
	var (
		timer       = uint64(5) // diff 5s
		lastMsg     uint64
		lastRoomMsg uint64
	)
	for {
		s.SpeedMsgSecond = (atomic.LoadUint64(&s.AllMsg) - lastMsg) / timer
		s.SpeedRoomBatchSecond = (atomic.LoadUint64(&s.BroadcastRoomMsg) - lastRoomMsg) / timer
		lastRoomMsg = s.BroadcastRoomMsg
		lastMsg = s.AllMsg
		time.Sleep(time.Duration(timer) * time.Second)
	}
}

func (s *Stat) IncrPushMsg() {
	atomic.AddUint64(&s.PushMsg, 1)
}

func (s *Stat) IncrBroadcastMsg() {
	atomic.AddUint64(&s.BroadcastMsg, 1)
}

func (s *Stat) IncrBroadcastRoomMsg() {
	atomic.AddUint64(&s.BroadcastRoomMsg, 1)
}

func (s *Stat) IncrPushMsgFailed() {
	atomic.AddUint64(&s.PushMsgFailed, 1)
}

func (s *Stat) IncrBroadcastMsgFailed() {
	atomic.AddUint64(&s.BroadcastMsgFailed, 1)
}

func (s *Stat) IncrBroadcastRoomMsgFailed() {
	atomic.AddUint64(&s.BroadcastRoomMsgFailed, 1)
}

func (s *Stat) IncrAllMsg() {
	atomic.AddUint64(&s.AllMsg, 1)
}

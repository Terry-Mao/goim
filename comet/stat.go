package main

import (
	"sync/atomic"
	"time"
)

type Stat struct {
	// online
	TcpOnline int64 `json:"tcp_online"`
	WsOnline  int64 `json:"websocket_online"`
	// messages
	AllMsg           uint64 `json:"all_msg"`
	PushMsg          uint64 `json:"push_msg"`
	BroadcastMsg     uint64 `json:"broadcast_msg"`
	BroadcastRoomMsg uint64 `json:"broadcast_room_msg"`
	// speed
	SpeedMsgSecond uint64 `json:"speed_msg_second"`
	// buckets
	BucketChannels map[int]int `json:"bucket_channels"`
	BucketRooms    map[int]int `json:"bucket_rooms"`
}

func NewStat() *Stat {
	s := &Stat{BucketChannels: make(map[int]int), BucketRooms: make(map[int]int)}
	go s.procSpeed()
	return s
}

func (s *Stat) Info() *Stat {
	var (
		idx    int
		bucket *Bucket
	)
	for idx, bucket = range DefaultServer.Buckets {
		s.BucketChannels[idx] = bucket.RoomCount()
		s.BucketRooms[idx] = bucket.ChannelCount()
	}
	return s
}

func (s *Stat) Reset() {
	atomic.StoreInt64(&s.TcpOnline, 0)
	atomic.StoreInt64(&s.WsOnline, 0)
	atomic.StoreUint64(&s.AllMsg, 0)
	atomic.StoreUint64(&s.PushMsg, 0)
	atomic.StoreUint64(&s.BroadcastMsg, 0)
	atomic.StoreUint64(&s.BroadcastRoomMsg, 0)
}

func (s *Stat) procSpeed() {
	var (
		timer   = uint64(5) // diff 5s
		lastMsg uint64
	)
	for {
		s.SpeedMsgSecond = (atomic.LoadUint64(&s.AllMsg) - lastMsg) / timer
		lastMsg = s.AllMsg
		time.Sleep(time.Duration(timer) * time.Second)
	}
}

func (s *Stat) IncrTcpOnline() {
	atomic.AddInt64(&s.TcpOnline, 1)
}

func (s *Stat) DecrTcpOnline() {
	atomic.AddInt64(&s.TcpOnline, -1)
}

func (s *Stat) IncrWsOnline() {
	atomic.AddInt64(&s.WsOnline, 1)
}

func (s *Stat) DecrWsOnline() {
	atomic.AddInt64(&s.WsOnline, -1)
}

func (s *Stat) IncrPushMsg() {
	atomic.AddUint64(&s.PushMsg, 1)
	atomic.AddUint64(&s.AllMsg, 1)
}

func (s *Stat) IncrBroadcastMsg() {
	atomic.AddUint64(&s.BroadcastMsg, 1)
	atomic.AddUint64(&s.AllMsg, 1)
}

func (s *Stat) IncrBroadcastRoomMsg() {
	atomic.AddUint64(&s.BroadcastRoomMsg, 1)
	atomic.AddUint64(&s.AllMsg, 1)
}

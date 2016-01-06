package main

import (
	log "code.google.com/p/log4go"
	itime "github.com/Terry-Mao/goim/libs/time"
	"sync"
	"time"
)

type RoomOptions struct {
	ChannelSize int
	BatchNum    int
	SignalTime  time.Duration
}

type Room struct {
	id     int32
	rLock  sync.RWMutex
	signal chan int
	// TODO use double-linked list
	chs map[*Channel]struct{} // map room id with channels
	// buffer msg
	mn   int
	n    int
	vers []int32
	ops  []int32
	msgs [][]byte
}

// NewRoom new a room struct, store channel room info.
func NewRoom(id int32, t *itime.Timer, options RoomOptions) (r *Room) {
	r = new(Room)
	r.id = id
	r.signal = make(chan int, SignalNum)
	r.chs = make(map[*Channel]struct{}, options.ChannelSize)
	r.n = 0
	r.mn = options.BatchNum
	r.vers = make([]int32, options.BatchNum)
	r.ops = make([]int32, options.BatchNum)
	r.msgs = make([][]byte, options.BatchNum)
	go r.push(t, options.BatchNum, options.SignalTime)
	return
}

// Put put channel into the room.
func (r *Room) Put(ch *Channel) {
	r.rLock.Lock()
	r.chs[ch] = struct{}{}
	r.rLock.Unlock()
	return
}

// Del delete channel from the room.
func (r *Room) Del(ch *Channel) {
	r.rLock.Lock()
	delete(r.chs, ch)
	r.rLock.Unlock()
}

// push merge proto and push msgs in batch.
func (r *Room) push(timer *itime.Timer, batch int, sigTime time.Duration) {
	var (
		done  bool
		ch    *Channel
		last  time.Time
		td    *itime.TimerData
		least time.Duration
	)
	if Debug {
		log.Debug("start room: %d goroutine", r.id)
	}
	td = timer.Add(sigTime, func() {
		select {
		case r.signal <- ProtoReady:
		default:
		}
	})
	for {
		if r.n > 0 {
			if least = sigTime - time.Now().Sub(last); least > 0 {
				timer.Set(td, least)
			} else {
				// timeout
				done = true
			}
		} else {
			last = time.Now()
		}
		if !done {
			if <-r.signal != ProtoReady {
				break
			}
			// try merge msg hard
			if r.n < batch {
				continue
			}
		}
		r.rLock.RLock()
		for ch, _ = range r.chs {
			ch.Pushs(r.vers[:r.n], r.ops[:r.n], r.msgs[:r.n])
		}
		// r.n set here and Push, so RLock is exclusive with Lock
		r.n = 0
		r.rLock.RUnlock()
		done = false
	}
	timer.Del(td)
	if Debug {
		log.Debug("room: %d goroutine exit", r.id)
	}
}

// Push push msg to the room.
func (r *Room) Push(ver int16, operation int32, msg []byte) (err error) {
	r.rLock.Lock()
	if r.n < r.mn {
		r.vers[r.n] = int32(ver)
		r.ops[r.n] = operation
		r.msgs[r.n] = msg
		r.n++
	}
	r.rLock.Unlock()
	select {
	case r.signal <- ProtoReady:
	default:
	}
	return
}

// Online get online number.
func (r *Room) Online() (o int) {
	r.rLock.RLock()
	o = len(r.chs)
	r.rLock.RUnlock()
	return
}

// Close close the room.
func (r *Room) Close() {
	var ch *Channel
	// if chan full, wait
	r.signal <- ProtoFinish
	r.rLock.RLock()
	for ch, _ = range r.chs {
		ch.Close()
	}
	r.rLock.RUnlock()
}

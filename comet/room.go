package main

import (
	log "code.google.com/p/log4go"
	"sync"
	"time"
)

type RoomOptions struct {
	ChannelSize int
	ProtoSize   int
	BatchNum    int
	SignalTime  time.Duration
}

type Room struct {
	id      int32
	rLock   sync.RWMutex
	proto   Ring
	signal  chan int
	chs     map[*Channel]struct{} // map room id with channels
	timer   *time.Timer
	sigTime time.Duration
	options RoomOptions
}

// NewRoom new a room struct, store channel room info.
func NewRoom(id int32, options RoomOptions) (r *Room) {
	r = new(Room)
	r.id = id
	r.options = options
	r.proto.Init(options.ProtoSize)
	r.signal = make(chan int, SignalNum)
	r.chs = make(map[*Channel]struct{}, options.ChannelSize)
	r.timer = time.NewTimer(options.SignalTime)
	go r.push()
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
func (r *Room) push() {
	var (
		s, n       int
		done       bool
		err        error
		p          *Proto
		ch         *Channel
		least      time.Duration
		last       = time.Now()
		vers       = make([]int32, r.options.BatchNum)
		operations = make([]int32, r.options.BatchNum)
		msgs       = make([][]byte, r.options.BatchNum)
	)
	if Debug {
		log.Debug("start room: %d goroutine", r.id)
	}
	for {
		if least = r.options.SignalTime - time.Now().Sub(last); least > 0 {
			r.timer.Reset(least)
			select {
			case s = <-r.signal:
				if s != ProtoReady {
					goto failed
				}
				// merge msgs
				if n == 0 {
					last = time.Now()
				}
				for {
					if n >= r.options.BatchNum {
						// msgs full
						done = true
						break
					}
					if p, err = r.proto.Get(); err != nil {
						// must be empty error
						break
					}
					vers[n] = int32(p.Ver)
					operations[n] = p.Operation
					msgs[n] = p.Body
					n++
					r.proto.GetAdv()
				}
			case <-r.timer.C:
				done = true
			}
		} else {
			done = true
		}
		if !done {
			continue
		}
		if n > 0 {
			r.rLock.RLock()
			for ch, _ = range r.chs {
				// ignore error
				ch.PushMsgs(vers[:n], operations[:n], msgs[:n])
			}
			r.rLock.RUnlock()
			n = 0
		}
		done = false
		last = time.Now()
	}
failed:
	r.timer.Stop()
	if Debug {
		log.Debug("room: %d goroutine exit", r.id)
	}
}

// PushMsg push msg to the room.
func (r *Room) PushMsg(ver int16, operation int32, msg []byte) (err error) {
	var proto *Proto
	r.rLock.Lock()
	if proto, err = r.proto.Set(); err == nil {
		proto.Ver = ver
		proto.Operation = operation
		proto.Body = msg
		r.proto.SetAdv()
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

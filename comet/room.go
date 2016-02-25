package main

import (
	log "code.google.com/p/log4go"
	"github.com/Terry-Mao/goim/libs/bytes"
	"github.com/Terry-Mao/goim/libs/define"
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
	id    int32
	rLock sync.RWMutex
	// map room id with channels
	// TODO use double-linked list
	chs   map[*Channel]struct{}
	proto chan *Proto
}

var (
	roomReadyProto = &Proto{Operation: define.OP_ROOM_READY}
)

// NewRoom new a room struct, store channel room info.
func NewRoom(id int32, t *itime.Timer, options RoomOptions) (r *Room) {
	r = new(Room)
	r.id = id
	r.proto = make(chan *Proto, options.BatchNum*2)
	r.chs = make(map[*Channel]struct{}, options.ChannelSize)
	go r.pushproc(t, options.BatchNum, options.SignalTime)
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

// Push push msg to the room, if chan full discard it.
func (r *Room) Push(ver int16, operation int32, msg []byte) (err error) {
	var p = &Proto{Ver: ver, Operation: operation, Body: msg}
	select {
	case r.proto <- p:
	default:
		err = ErrRoomFull
	}
	return
}

// EPush ensure push msg to the room.
func (r *Room) EPush(ver int16, operation int32, msg []byte) (err error) {
	var p = &Proto{Ver: ver, Operation: operation, Body: msg}
	r.proto <- p
	return
}

// Online get online number.
func (r *Room) Online() (o int) {
	r.rLock.RLock()
	o = len(r.chs)
	r.rLock.RUnlock()
	return
}

// pushproc merge proto and push msgs in batch.
func (r *Room) pushproc(timer *itime.Timer, batch int, sigTime time.Duration) {
	var (
		n     int
		done  bool
		last  time.Time
		least time.Duration
		p     *Proto
		ch    *Channel
		td    *itime.TimerData
		buf   = bytes.NewWriterSize(int(MaxBodySize))
	)
	if Debug {
		log.Debug("start room: %d goroutine", r.id)
	}
	td = timer.Add(sigTime, func() {
		select {
		case r.proto <- roomReadyProto:
		default:
		}
	})
	for {
		if n > 0 {
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
			if p = <-r.proto; p == nil {
				break // exit
			} else if p != roomReadyProto {
				// merge buffer ignore error, always nil
				p.WriteTo(buf)
				if n++; n < batch {
					continue
				}
			}
		}
		r.rLock.RLock()
		for ch, _ = range r.chs {
			ch.Push(0, define.OP_RAW, buf.Buffer())
		}
		r.rLock.RUnlock()
		n = 0
		done = false
		ch = nil // avoid gc memory leak
		// after push to room channel, renew a buffer, let old buffer gc
		buf = bytes.NewWriterSize(buf.Size())
	}
	timer.Del(td)
	if Debug {
		log.Debug("room: %d goroutine exit", r.id)
	}
}

// Close close the room.
func (r *Room) Close() {
	var ch *Channel
	r.proto <- nil
	r.rLock.RLock()
	for ch, _ = range r.chs {
		ch.Close()
	}
	r.rLock.RUnlock()
}

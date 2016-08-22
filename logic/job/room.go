package main

import (
	"goim/libs/bytes"
	"goim/libs/define"
	"goim/libs/proto"
	itime "goim/libs/time"
	"sync"
	"time"

	log "github.com/thinkboy/log4go"
)

const (
	roomMapCup = 100
)

var roomBucket *RoomBucket

type RoomBucket struct {
	roomNum int
	rooms   map[int32]*Room
	rwLock  sync.RWMutex
	options RoomOptions
	round   *Round
}

func InitRoomBucket(r *Round, options RoomOptions) {
	roomBucket = &RoomBucket{
		roomNum: 0,
		rooms:   make(map[int32]*Room, roomMapCup),
		rwLock:  sync.RWMutex{},
		options: options,
		round:   r,
	}
}

func (this *RoomBucket) Get(roomId int32) (r *Room) {
	this.rwLock.Lock()
	room, ok := this.rooms[roomId]
	if !ok {
		room = NewRoom(roomId, this.round.Timer(this.roomNum), this.options)
		this.rooms[roomId] = room
		this.roomNum++
		log.Debug("new roomId:%d num:%d", roomId, this.roomNum)
	}
	this.rwLock.Unlock()
	return room
}

type RoomOptions struct {
	BatchNum   int
	SignalTime time.Duration
}

type Room struct {
	id    int32
	proto chan *proto.Proto
}

var (
	roomReadyProto = &proto.Proto{Operation: define.OP_ROOM_READY}
)

// NewRoom new a room struct, store channel room info.
func NewRoom(id int32, t *itime.Timer, options RoomOptions) (r *Room) {
	r = new(Room)
	r.id = id
	r.proto = make(chan *proto.Proto, options.BatchNum*2)
	go r.pushproc(t, options.BatchNum, options.SignalTime)
	return
}

// Push push msg to the room, if chan full discard it.
func (r *Room) Push(ver int16, operation int32, msg []byte) (err error) {
	var p = &proto.Proto{Ver: ver, Operation: operation, Body: msg}
	select {
	case r.proto <- p:
	default:
		err = ErrRoomFull
	}
	return
}

// EPush ensure push msg to the room.
func (r *Room) EPush(ver int16, operation int32, msg []byte) {
	var p = &proto.Proto{Ver: ver, Operation: operation, Body: msg}
	r.proto <- p
	return
}

// pushproc merge proto and push msgs in batch.
func (r *Room) pushproc(timer *itime.Timer, batch int, sigTime time.Duration) {
	var (
		n    int
		last time.Time
		p    *proto.Proto
		td   *itime.TimerData
		buf  = bytes.NewWriterSize(int(proto.MaxBodySize))
	)
	log.Debug("start room: %d goroutine", r.id)
	td = timer.Add(sigTime, func() {
		select {
		case r.proto <- roomReadyProto:
		default:
		}
	})
	for {
		if p = <-r.proto; p == nil {
			break // exit
		} else if p != roomReadyProto {
			// merge buffer ignore error, always nil
			p.WriteTo(buf)
			if n++; n == 1 {
				last = time.Now()
				timer.Set(td, sigTime)
				continue
			} else if n < batch {
				if sigTime > time.Now().Sub(last) {
					continue
				}
			}
		} else {
			if n == 0 {
				continue
			}
		}
		broadcastRoomBytes(r.id, buf.Buffer())
		n = 0
		// TODO use reset buffer
		// after push to room channel, renew a buffer, let old buffer gc
		buf = bytes.NewWriterSize(buf.Size())
	}
	timer.Del(td)
	log.Debug("room: %d goroutine exit", r.id)
}

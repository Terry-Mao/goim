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
	bLock   sync.RWMutex
	options RoomOptions
	round   *Round
}

func InitRoomBucket(r *Round, options RoomOptions) {
	roomBucket = &RoomBucket{
		roomNum: 0,
		rooms:   make(map[int32]*Room, roomMapCup),
		bLock:   sync.RWMutex{},
		options: options,
		round:   r,
	}
}

func (b *RoomBucket) Get(roomId int32) (r *Room) {
	b.bLock.Lock()
	room, ok := b.rooms[roomId]
	if !ok {
		room = NewRoom(roomId, b.round.Timer(b.roomNum), b.options)
		b.rooms[roomId] = room
		b.roomNum++
		log.Debug("new roomId:%d num:%d", roomId, b.roomNum)
	}
	b.bLock.Unlock()
	return room
}

func (b *RoomBucket) Del(roomId int32) {
	b.bLock.Lock()
	delete(b.rooms, roomId)
	b.bLock.Unlock()
}

type RoomOptions struct {
	BatchNum   int
	SignalTime time.Duration
	IdleTime   time.Duration
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
	go r.pushproc(t, options.BatchNum, options.SignalTime, options.IdleTime)
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
func (r *Room) pushproc(timer *itime.Timer, batch int, sigTime time.Duration, idleTime time.Duration) {
	var (
		n   int
		p   *proto.Proto
		td  *itime.TimerData
		buf = bytes.NewWriterSize(int(proto.MaxBodySize))
	)
	log.Debug("start room: %d goroutine", r.id)
	td = timer.Add(idleTime, func() {
		select {
		case r.proto <- roomReadyProto:
		default:
		}
	})
	for {
		if p = <-r.proto; p != roomReadyProto {
			// merge buffer ignore error, always nil
			p.WriteTo(buf)
			// batch
			if n++; n == 1 {
				timer.Set(td, sigTime)
				continue
			} else if n < batch {
				continue
			}
		} else if n == 0 {
			// idle
			break
		}
		timer.Set(td, idleTime)
		broadcastRoomBytes(r.id, buf.Buffer())
		// TODO use reset buffer
		// after push to room channel, renew a buffer, let old buffer gc
		buf = bytes.NewWriterSize(buf.Size())
		n = 0
	}
	timer.Del(td)
	roomBucket.Del(r.id)
	log.Debug("end room: %d goroutine exit", r.id)
}

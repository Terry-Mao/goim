package job

import (
	"errors"
	"fmt"
	"time"

	comet "github.com/Terry-Mao/goim/api/comet/grpc"
	"github.com/Terry-Mao/goim/internal/job/conf"
	"github.com/Terry-Mao/goim/pkg/bytes"
	log "github.com/golang/glog"
)

var (
	// ErrComet commet error.
	ErrComet = errors.New("comet rpc is not available")

	// ErrCometFull comet chan full.
	ErrCometFull = errors.New("comet proto chan full")

	// ErrRoomFull room chan full.
	ErrRoomFull = errors.New("room proto chan full")

	// 通知該推送房間消息的flag
	roomReadyProto = new(comet.Proto)
)

// Room room.
type Room struct {
	c *conf.Room

	//
	job *Job

	//
	id string

	//
	proto chan *comet.Proto
}

// NewRoom new a room struct, store channel room info.
func NewRoom(job *Job, id string, c *conf.Room) (r *Room) {
	r = &Room{
		c:     c,
		id:    id,
		job:   job,
		proto: make(chan *comet.Proto, c.Batch*2),
	}
	go r.pushproc(c.Batch, time.Duration(c.Signal))
	return
}

// Push push msg to the room, if chan full discard it.
func (r *Room) Push(op int32, msg []byte) (err error) {
	var p = &comet.Proto{
		Ver:  1,
		Op:   op,
		Body: msg,
	}
	select {
	case r.proto <- p:
	default:
		err = ErrRoomFull
	}
	return
}

// 房間消息聚合
func (r *Room) pushproc(batch int, sigTime time.Duration) {
	var (
		// 緩衝的訊息筆數
		n int

		// 第一筆開始緩衝的時間
		last time.Time

		//
		p *comet.Proto

		// 緩衝訊息
		buf = bytes.NewWriterSize(int(comet.MaxBodySize))
	)
	fmt.Println("new")
	log.Infof("start room:%s goroutine", r.id)

	// 控制多久才推送訊息給comet server
	td := time.AfterFunc(sigTime, func() {
		select {
		case r.proto <- roomReadyProto:
		default:
		}
	})

	defer td.Stop()
	//
	for {
		if p = <-r.proto; p == nil {
			break // exit
		} else if p != roomReadyProto {
			p.WriteTo(buf)

			// 每次緩衝的第一筆都需重置多久後才推送，因為此goroutine雖然在啟動時AfterFunc一次時間
			// 但這只是為了設定任務，倒數時間還是要由每次緩衝的第一筆來設置才會正確，因為此goroutine會一直
			// 運行下去，如果指運行一次確實不需要做time.Reset。
			// 紀錄第一次緩衝時間是為了每次接到訊息要檢查本次是否超過緩衝的時間，應該要推送給comet
			// 假設接到第一筆訊息，時間為2019-01-01 00:00:00，之後會有兩種情況出現
			// 1. 任務倒數未到等待下一筆
			//    (1) 未超過最大緩衝筆數
			//        比對至從緩衝第一筆到這筆之間是否已超過要推送給comet的時間差，如果還沒到就繼續等待下筆或是時間到
			//    (2) 超過最大緩衝筆數
			//        執行訊息推送給comet
			// 2. 任務倒數已到
			//    (1) 有緩衝到一筆就推送給comet
			//    (2) 如果沒緩衝到一筆代表這個房間都沒人推送了，就可以close goroutine
			if n++; n == 1 {
				last = time.Now()
				td.Reset(sigTime)
				continue
			} else if n < batch {
				if sigTime > time.Since(last) {
					continue
				}
			}
		} else {
			if n == 0 {
				fmt.Println("close")
				break
			}
		}

		_ = r.job.broadcastRoomRawBytes(r.id, buf.Buffer())
		// TODO use reset buffer
		// after push to room channel, renew a buffer, let old buffer gc
		buf = bytes.NewWriterSize(buf.Size())
		n = 0
		if r.c.Idle != 0 {
			td.Reset(time.Duration(r.c.Idle))
		} else {
			td.Reset(time.Minute)
		}
	}
	r.job.delRoom(r.id)
	log.Infof("room:%s goroutine exit", r.id)
}

func (j *Job) delRoom(roomID string) {
	j.roomsMutex.Lock()
	delete(j.rooms, roomID)
	j.roomsMutex.Unlock()
}

// 根據room id取Room
func (j *Job) getRoom(roomID string) *Room {
	j.roomsMutex.RLock()
	room, ok := j.rooms[roomID]
	j.roomsMutex.RUnlock()
	if !ok {
		j.roomsMutex.Lock()
		if room, ok = j.rooms[roomID]; !ok {
			room = NewRoom(j, roomID, j.c.Room)
			j.rooms[roomID] = room
		}
		j.roomsMutex.Unlock()
		log.Infof("new a room:%s active:%d", roomID, len(j.rooms))
	}
	return room
}

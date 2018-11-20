package job

import (
	"context"
	"fmt"
	"net/url"
	"sync/atomic"
	"time"

	"github.com/Bilibili/discovery/naming"
	pb "github.com/Terry-Mao/goim/api/comet/grpc"
	"github.com/Terry-Mao/goim/internal/job/conf"

	log "github.com/golang/glog"
	"google.golang.org/grpc"
)

func newCometClient(addr string) (pb.CometClient, error) {
	opts := []grpc.DialOption{
		// grpc.WithBlock(),
		grpc.WithInsecure(),
		grpc.WithTimeout(time.Second),
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second))
	defer cancel()
	conn, err := grpc.DialContext(ctx, addr, opts...)
	if err != nil {
		return nil, err
	}
	return pb.NewCometClient(conn), err
}

// Comet is a comet.
type Comet struct {
	serverID      string
	client        pb.CometClient
	pushChan      []chan *pb.PushMsgReq
	roomChan      []chan *pb.BroadcastRoomReq
	broadcastChan chan *pb.BroadcastReq
	pushChanNum   uint64
	roomChanNum   uint64
	routineSize   uint64

	ctx    context.Context
	cancel context.CancelFunc
}

// NewComet new a comet.
func NewComet(in *naming.Instance, c *conf.Comet) (*Comet, error) {
	cmt := &Comet{
		serverID:      in.Hostname,
		pushChan:      make([]chan *pb.PushMsgReq, c.RoutineSize),
		roomChan:      make([]chan *pb.BroadcastRoomReq, c.RoutineSize),
		broadcastChan: make(chan *pb.BroadcastReq, c.RoutineSize),
		routineSize:   uint64(c.RoutineSize),
	}
	var grpcAddr string
	for _, addrs := range in.Addrs {
		u, err := url.Parse(addrs)
		if err == nil && u.Scheme == "grpc" {
			grpcAddr = u.Host
		}
	}
	if grpcAddr == "" {
		return nil, fmt.Errorf("invalid grpc address:%v", in.Addrs)
	}
	var err error
	if cmt.client, err = newCometClient(grpcAddr); err != nil {
		return nil, err
	}
	cmt.ctx, cmt.cancel = context.WithCancel(context.Background())

	for i := 0; i < c.RoutineSize; i++ {
		cmt.pushChan[i] = make(chan *pb.PushMsgReq, c.RoutineChan)
		cmt.roomChan[i] = make(chan *pb.BroadcastRoomReq, c.RoutineChan)
		go cmt.process(cmt.pushChan[i], cmt.roomChan[i], cmt.broadcastChan)
	}
	return cmt, nil
}

// Push push a user message.
func (c *Comet) Push(arg *pb.PushMsgReq) (err error) {
	idx := atomic.AddUint64(&c.pushChanNum, 1) % c.routineSize
	c.pushChan[idx] <- arg
	return
}

// BroadcastRoom broadcast a room message.
func (c *Comet) BroadcastRoom(arg *pb.BroadcastRoomReq) (err error) {
	idx := atomic.AddUint64(&c.roomChanNum, 1) % c.routineSize
	c.roomChan[idx] <- arg
	return
}

// Broadcast broadcast a message.
func (c *Comet) Broadcast(arg *pb.BroadcastReq) (err error) {
	c.broadcastChan <- arg
	return
}

func (c *Comet) process(pushChan chan *pb.PushMsgReq, roomChan chan *pb.BroadcastRoomReq, broadcastChan chan *pb.BroadcastReq) {
	var err error
	for {
		select {
		case broadcastArg := <-broadcastChan:
			_, err = c.client.Broadcast(context.Background(), &pb.BroadcastReq{
				Proto:    broadcastArg.Proto,
				ProtoOp:  broadcastArg.ProtoOp,
				Speed:    broadcastArg.Speed,
				Platform: broadcastArg.Platform,
			})
			if err != nil {
				log.Errorf("c.client.Broadcast(%s, reply) serverId:%s error(%v)", broadcastArg, c.serverID, err)
			}
		case roomArg := <-roomChan:
			_, err = c.client.BroadcastRoom(context.Background(), &pb.BroadcastRoomReq{
				RoomID: roomArg.RoomID,
				Proto:  roomArg.Proto,
			})
			if err != nil {
				log.Errorf("c.client.BroadcastRoom(%s, reply) serverId:%s error(%v)", roomArg, c.serverID, err)
			}
		case pushArg := <-pushChan:
			_, err = c.client.PushMsg(context.Background(), &pb.PushMsgReq{
				Keys:    pushArg.Keys,
				Proto:   pushArg.Proto,
				ProtoOp: pushArg.ProtoOp,
			})
			if err != nil {
				log.Errorf("c.client.PushMsg(%s, reply) serverId:%s error(%v)", pushArg, c.serverID, err)
			}
		case <-c.ctx.Done():
			return
		}
	}
}

// Close close the resouces.
func (c *Comet) Close() (err error) {
	finish := make(chan bool)
	go func() {
		for {
			n := len(c.broadcastChan)
			for _, ch := range c.pushChan {
				n += len(ch)
			}
			for _, ch := range c.roomChan {
				n += len(ch)
			}
			if n == 0 {
				finish <- true
				return
			}
			time.Sleep(time.Second)
		}
	}()
	select {
	case <-finish:
		log.Info("close comet finish")
	case <-time.After(5 * time.Second):
		err = fmt.Errorf("close comet(server:%s push:%d room:%d broadcast:%d) timeout", c.serverID, len(c.pushChan), len(c.roomChan), len(c.broadcastChan))
	}
	c.cancel()
	return
}

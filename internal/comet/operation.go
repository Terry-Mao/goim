package comet

import (
	"context"

	model "github.com/Terry-Mao/goim/api/comet/grpc"
	logic "github.com/Terry-Mao/goim/api/logic/grpc"
	"github.com/Terry-Mao/goim/internal/comet/errors"
	"github.com/Terry-Mao/goim/pkg/strings"
	log "github.com/golang/glog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding/gzip"
)

// Connect connected a connection.
func (s *Server) Connect(c context.Context, p *model.Proto, cookie string) (mid int64, key, rid string, tags []string, accepts []int32, err error) {
	reply, err := s.rpcClient.Connect(c, &logic.ConnectReq{
		Server:    s.serverID,
		ServerKey: s.NextKey(),
		Cookie:    cookie,
		Token:     p.Body,
	})
	if err != nil {
		return
	}
	return reply.Mid, reply.Key, reply.RoomID, reply.Tags, reply.Accepts, nil
}

// Disconnect disconnected a connection.
func (s *Server) Disconnect(c context.Context, mid int64, key string) (err error) {
	_, err = s.rpcClient.Disconnect(context.Background(), &logic.DisconnectReq{
		Server: s.serverID,
		Mid:    mid,
		Key:    key,
	})
	return
}

// Heartbeat heartbeat a connection session.
func (s *Server) Heartbeat(ctx context.Context, mid int64, key string) (err error) {
	_, err = s.rpcClient.Heartbeat(ctx, &logic.HeartbeatReq{
		Server: s.serverID,
		Mid:    mid,
		Key:    key,
	})
	return
}

// RenewOnline renew room online.
func (s *Server) RenewOnline(ctx context.Context, serverID string, rommCount map[string]int32) (allRoom map[string]int32, err error) {
	reply, err := s.rpcClient.RenewOnline(ctx, &logic.OnlineReq{
		Server:    s.serverID,
		RoomCount: rommCount,
	}, grpc.UseCompressor(gzip.Name))
	if err != nil {
		return
	}
	return reply.AllRoomCount, nil
}

// Operate operate.
func (s *Server) Operate(c context.Context, p *model.Proto, ch *Channel, b *Bucket) (err error) {
	switch {
	case p.Op >= model.MinBusinessOp && p.Op <= model.MaxBusinessOp:
		_, err1 := s.rpcClient.Receive(c, &logic.ReceiveReq{Mid: ch.Mid, Proto: p})
		if err1 != nil {
			// TODO ack failed
			log.Errorf("s.rpcClient.Receive operation:%d error(%v)", p.Op, err)
		}
		// TODO ack ok
		p.Body = nil
	case p.Op == model.OpChangeRoom:
		err = b.ChangeRoom(string(p.Body), ch)
		p.Op = model.OpChangeRoomReply
	case p.Op == model.OpRegister:
		if ops, err := strings.SplitInt32s(string(p.Body), ","); err == nil {
			ch.Watch(ops...)
		}
		p.Op = model.OpRegisterReply
	case p.Op == model.OpUnregister:
		if ops, err := strings.SplitInt32s(string(p.Body), ","); err == nil {
			ch.UnWatch(ops...)
		}
		p.Op = model.OpUnregisterReply
	default:
		err = errors.ErrOperation
	}
	if err != nil {
		log.Errorf("operate(%+v) error(%v)", p, err)
	}
	return
}

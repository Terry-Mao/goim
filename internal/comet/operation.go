package comet

import (
	"context"
	"time"

	"github.com/Terry-Mao/goim/api/logic"
	"github.com/Terry-Mao/goim/api/protocol"
	"github.com/Terry-Mao/goim/pkg/strings"
	log "github.com/golang/glog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding/gzip"
)

// Connect connected a connection.
func (s *Server) Connect(c context.Context, p *protocol.Proto, cookie string) (mid int64, key, rid string, accepts []int32, heartbeat time.Duration, err error) {
	reply, err := s.rpcClient.Connect(c, &logic.ConnectReq{
		Server: s.serverID,
		Cookie: cookie,
		Token:  p.Body,
	})
	if err != nil {
		return
	}
	return reply.Mid, reply.Key, reply.RoomID, reply.Accepts, time.Duration(reply.Heartbeat), nil
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
func (s *Server) RenewOnline(ctx context.Context, serverID string, roomCount map[string]int32) (allRoom map[string]int32, err error) {
	reply, err := s.rpcClient.RenewOnline(ctx, &logic.OnlineReq{
		Server:    s.serverID,
		RoomCount: roomCount,
	}, grpc.UseCompressor(gzip.Name))
	if err != nil {
		return
	}
	return reply.AllRoomCount, nil
}

// Receive receive a message.
func (s *Server) Receive(ctx context.Context, mid int64, p *protocol.Proto) (err error) {
	_, err = s.rpcClient.Receive(ctx, &logic.ReceiveReq{Mid: mid, Proto: p})
	return
}

// Operate operate.
func (s *Server) Operate(ctx context.Context, p *protocol.Proto, ch *Channel, b *Bucket) error {
	switch p.Op {
	case protocol.OpChangeRoom:
		if err := b.ChangeRoom(string(p.Body), ch); err != nil {
			log.Errorf("b.ChangeRoom(%s) error(%v)", p.Body, err)
		}
		p.Op = protocol.OpChangeRoomReply
	case protocol.OpSub:
		if ops, err := strings.SplitInt32s(string(p.Body), ","); err == nil {
			ch.Watch(ops...)
		}
		p.Op = protocol.OpSubReply
	case protocol.OpUnsub:
		if ops, err := strings.SplitInt32s(string(p.Body), ","); err == nil {
			ch.UnWatch(ops...)
		}
		p.Op = protocol.OpUnsubReply
	default:
		// TODO ack ok&failed
		if err := s.Receive(ctx, ch.Mid, p); err != nil {
			log.Errorf("s.Report(%d) op:%d error(%v)", ch.Mid, p.Op, err)
		}
		p.Body = nil
	}
	return nil
}

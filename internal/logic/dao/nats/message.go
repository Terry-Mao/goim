package nats

import (
	"context"
	"strconv"

	"github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"
	"github.com/nats-io/go-nats"

	pb "github.com/Terry-Mao/goim/api/logic/grpc"
	"github.com/Terry-Mao/goim/internal/logic/conf"
)

// Dao dao for nats
type Dao struct {
	c           *conf.Config
	natsClient  *nats.Conn
	redis       *redis.Pool
	redisExpire int32
}

// NatsDao alias name of dao
type NatsDao = Dao

// LogicConfig configuration for nats / liftbridge queue
type Config struct {
	Channel   string
	ChannelID string
	Group     string
	NatsAddr  string
	LiftAddr  string
}

// NatsCOnfig alias name of Config
type NatsConfig = Config

// PushMsg push a message to databus.
func (d *Dao) PushMsg(c context.Context, op int32, server string, keys []string, msg []byte) (err error) {
	pushMsg := &pb.PushMsg{
		Type:      pb.PushMsg_PUSH,
		Operation: op,
		Server:    server,
		Keys:      keys,
		Msg:       msg,
	}
	b, err := proto.Marshal(pushMsg)
	if err != nil {
		return
	}

	_ = d.publishMessage(d.c.Nats.Topic, d.c.Nats.AckInbox, []byte(keys[0]), b)
	return
}

// BroadcastRoomMsg push a message to databus.
func (d *Dao) BroadcastRoomMsg(c context.Context, op int32, room string, msg []byte) (err error) {
	pushMsg := &pb.PushMsg{
		Type:      pb.PushMsg_ROOM,
		Operation: op,
		Room:      room,
		Msg:       msg,
	}
	b, err := proto.Marshal(pushMsg)
	if err != nil {
		return
	}

	_ = d.publishMessage(d.c.Nats.Topic, d.c.Nats.AckInbox, []byte(room), b)
	return
}

// BroadcastMsg push a message to databus.
func (d *Dao) BroadcastMsg(c context.Context, op, speed int32, msg []byte) (err error) {
	pushMsg := &pb.PushMsg{
		Type:      pb.PushMsg_BROADCAST,
		Operation: op,
		Speed:     speed,
		Msg:       msg,
	}
	b, err := proto.Marshal(pushMsg)
	if err != nil {
		return
	}

	key := strconv.FormatInt(int64(op), 10)

	_ = d.publishMessage(d.c.Nats.Topic, d.c.Nats.AckInbox, []byte(key), b)

	return
}

package nats

import (
	"context"
	"time"

	"github.com/nats-io/go-nats"

	"github.com/Terry-Mao/goim/internal/logic/conf"
)

// New new a dao and return.
func New(c *conf.Config) *Dao {

	conn, err := newNatsClient(c.Nats.Brokers, c.Nats.Topic, c.Nats.TopicID)
	if err != nil {
		return nil
	}

	d := &Dao{
		c:          c,
		natsClient: conn,
		redis:      newRedis(c.Redis),
		// TODO: handler redis expire
		redisExpire: int32(time.Duration(c.Redis.Expire) / time.Second),
	}
	return d
}

// Close close the resource.
func (d *Dao) Close() error {
	if d.natsClient != nil {
		d.natsClient.Close()
	}

	return d.redis.Close()
}

// Ping dao ping.
func (d *Dao) Ping(c context.Context) error {
	return d.pingRedis(c)
}

func newNatsClient(natsAddr, channel, channelID string) (*nats.Conn, error) {

	// conn, err := nats.GetDefaultOptions().Connect()
	// natsAddr := "nats://localhost:4222"
	return nats.Connect(natsAddr)
}

func (d *Dao) publishMessage(channel, ackInbox string, key, value []byte) error {
	return d.natsClient.Publish(channel, value)
}

package dao

import (
	"errors"

	"github.com/nats-io/go-nats"

	"github.com/Terry-Mao/goim/internal/logic/conf"
)

// natsDao dao for nats
type natsDao struct {
	c    *conf.Config
	push *nats.Conn
}

// New new a dao and return.
func NewNats(c *conf.Config) *natsDao {

	conn, err := newNatsClient(c.Nats.Brokers, c.Nats.Topic, c.Nats.TopicID)
	if err != nil {
		return nil
	}
	d := &natsDao{
		c:    c,
		push: conn,
	}
	return d
}

// PublishMessage  push message to nats
func (d *natsDao) PublishMessage(topic, ackInbox string, key string, value []byte) error {
	if d.push == nil {
		return errors.New("nats error")
	}
	msg := &nats.Msg{Subject: topic, Reply: ackInbox, Data: value}
	return d.push.PublishMsg(msg)

}

// Close close the resource.
func (d *natsDao) Close() error {
	if d.push != nil {
		d.push.Close()
	}
	return nil
}

func newNatsClient(natsAddr, channel, channelID string) (*nats.Conn, error) {

	// conn, err := nats.GetDefaultOptions().Connect()
	// natsAddr := "nats://localhost:4222"
	return nats.Connect(natsAddr)
}

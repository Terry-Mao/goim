package nats

import (
	"context"
	"time"

	liftbridge "github.com/liftbridge-io/go-liftbridge"
	nats "github.com/nats-io/go-nats"

	"github.com/Terry-Mao/goim/internal/logic/conf"
)

// New new a dao and return.
func New(c *conf.LogicConfig) *Dao {

	conn, err := newNatsClient(c.Nats.NatsAddr, c.Nats.LiftAddr, c.Nats.Channel, c.Nats.ChannelID)
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
	d.natsClient.Close()
	return d.redis.Close()
}

// Ping dao ping.
func (d *Dao) Ping(c context.Context) error {
	return d.pingRedis(c)
}

func newNatsClient(natsAddr, liftAddr, channel, channelID string) (*nats.Conn, error) {
	// liftAddr := "localhost:9292" // address for lift-bridge
	// channel := "bar"
	// channelID := "bar-stream"
	// ackInbox := "acks"

	if err := createStream(liftAddr, channel, channelID); err != nil {
		if err != liftbridge.ErrStreamExists {
			return nil, err
		}
	}
	// conn, err := nats.GetDefaultOptions().Connect()
	// natsAddr := "nats://localhost:4222"
	return nats.Connect(natsAddr)

	// defer conn.Flush()
	// defer conn.Close()

}

func (d *Dao) publishMessage(channel, ackInbox string, key, value []byte) error {
	// var wg sync.WaitGroup
	// wg.Add(1)
	// sub, err := d.natsClient.Subscribe(ackInbox, func(m *nats.Msg) {
	// 	ack, err := liftbridge.UnmarshalAck(m.Data)
	// 	if err != nil {
	// 		// TODO: handel error write to log
	// 		return
	// 	}
	//
	// 	log.Info(utils.StrBuilder("ack:", ack.StreamSubject, " stream: ",  ack.StreamName, " offset: ",  strconv.FormatInt(ack.Offset,10), " msg: ",  ack.MsgSubject) )
	// 	wg.Done()
	// })
	// if err != nil {
	// 	return err
	// }
	// defer sub.Unsubscribe()

	m := liftbridge.NewMessage(value, liftbridge.MessageOptions{Key: key, AckInbox: ackInbox})

	if err := d.natsClient.Publish(channel, m); err != nil {
		return err
	}

	// wg.Wait()
	return nil
}

// func (d *Dao) publishMessageSync(channel, ackInbox string, key, value []byte) error {
// 	var wg sync.WaitGroup
// 	wg.Add(1)
// 	sub, err := d.natsClient.Subscribe(ackInbox, func(m *nats.Msg) {
// 		ack, err := liftbridge.UnmarshalAck(m.Data)
// 		if err != nil {
// 			// TODO: handel error write to log
// 			return
// 		}
//
// 		log.Info(utils.StrBuilder("ack:", ack.StreamSubject, " stream: ", ack.StreamName, " offset: ", strconv.FormatInt(ack.Offset, 10), " msg: ", ack.MsgSubject))
// 		wg.Done()
// 	})
// 	if err != nil {
// 		return err
// 	}
// 	defer sub.Unsubscribe()
//
// 	m := liftbridge.NewMessage(value, liftbridge.MessageOptions{Key: key, AckInbox: ackInbox})
//
// 	if err := d.natsClient.Publish(channel, m); err != nil {
// 		return err
// 	}
//
// 	wg.Wait()
// 	return nil
// }

func createStream(liftAddr, subject, name string) error {

	client, err := liftbridge.Connect([]string{liftAddr})
	if err != nil {
		return err
	}
	defer client.Close()

	stream := liftbridge.StreamInfo{
		Subject:           subject,
		Name:              name,
		ReplicationFactor: 1,
	}
	if err := client.CreateStream(context.Background(), stream); err != nil {
		if err != liftbridge.ErrStreamExists {
			return err
		}
	}

	return nil
}

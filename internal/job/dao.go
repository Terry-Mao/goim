package job

import (
	"context"

	cluster "github.com/bsm/sarama-cluster"
	"github.com/gogo/protobuf/proto"
	log "github.com/golang/glog"
	"github.com/nats-io/go-nats"

	pb "github.com/Terry-Mao/goim/api/logic/grpc"
	"github.com/Terry-Mao/goim/internal/job/conf"
)

type JobConsumer interface {
	Consume(j *Job)
	Close() error
}

type kafkaConsumer struct {
	consumer *cluster.Consumer
}
type natsConsumer struct {
	consumer *nats.Conn
}

func NewKafka(c *conf.Config) *kafkaConsumer {
	return &kafkaConsumer{
		consumer: newKafkaSub(c.Kafka),
	}
}

func (c *kafkaConsumer) Close() error {
	return c.consumer.Close()
}

// Consume messages, watch signals
func (c *kafkaConsumer) Consume(j *Job) {
	for {
		select {
		case err := <-c.consumer.Errors():
			log.Errorf("consumer error(%v)", err)
		case n := <-c.consumer.Notifications():
			log.Infof("consumer rebalanced(%v)", n)
		case msg, ok := <-c.consumer.Messages():
			if !ok {
				return
			}
			c.consumer.MarkOffset(msg, "")
			// process push message
			pushMsg := new(pb.PushMsg)
			if err := proto.Unmarshal(msg.Value, pushMsg); err != nil {
				log.Errorf("proto.Unmarshal(%v) error(%v)", msg, err)
				continue
			}
			if err := j.push(context.Background(), pushMsg); err != nil {
				log.Errorf("c.push(%v) error(%v)", pushMsg, err)
			}
			log.Infof("consume: %s/%d/%d\t%s\t%+v", msg.Topic, msg.Partition, msg.Offset, msg.Key, pushMsg)
		}
	}
}

// Consume messages, watch signals
func (c *natsConsumer) Consume(j *Job) {
	ctx := context.Background()

	// process push message
	pushMsg := new(pb.PushMsg)

	if _, err := c.consumer.Subscribe(j.c.Nats.Topic, func(msg *nats.Msg) {

		log.Info("------------> ", string(msg.Data))

		if err := proto.Unmarshal(msg.Data, pushMsg); err != nil {
			log.Errorf("proto.Unmarshal(%v) error(%v)", msg, err)
			return
		}
		if err := j.push(context.Background(), pushMsg); err != nil {
			log.Errorf("push(%v) error(%v)", pushMsg, err)
		}
		log.Infof("consume: %d  %s \t%+v", msg.Data, pushMsg)

	}); err != nil {
		return
	}

	<-ctx.Done()
	return
}
func NewNats(c *conf.Config) *natsConsumer {

	nc, err := nats.Connect(c.Nats.Brokers)
	if err != nil {
		return nil
	}

	return &natsConsumer{
		consumer: nc,
	}
}

func (c *natsConsumer) Close() error {
	c.consumer.Close()
	return nil
}

func newKafkaSub(c *conf.Kafka) *cluster.Consumer {
	config := cluster.NewConfig()
	config.Consumer.Return.Errors = true
	config.Group.Return.Notifications = true
	consumer, err := cluster.NewConsumer(c.Brokers, c.Group, []string{c.Topic}, config)
	if err != nil {
		panic(err)
	}
	return consumer
}

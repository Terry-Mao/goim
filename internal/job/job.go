package job

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/Bilibili/discovery/naming"
	"github.com/gogo/protobuf/proto"
	"github.com/nats-io/go-nats"

	pb "github.com/Terry-Mao/goim/api/logic/grpc"
	"github.com/Terry-Mao/goim/internal/job/conf"

	cluster "github.com/bsm/sarama-cluster"
	log "github.com/golang/glog"
)

// Job is push job.
type Job struct {
	c            *conf.Config
	consumer     JobConsumer
	cometServers map[string]*Comet

	rooms      map[string]*Room
	roomsMutex sync.RWMutex
}

type JobConsumer interface {
	// 	WatchComet(c *naming.Config)
	// Subscribe(channel, channelID string) error
	Consume(j *Job)
	Close() error
}

// New new a push job.
func New(c *conf.Config) *Job {
	j := &Job{
		c: c,
		// 	consumer: newKafkaSub(c.Kafka),
		rooms: make(map[string]*Room),
	}
	if c.UseNats {
		j.consumer = NewKafka(c)
	} else {
		j.consumer = NewKafka(c)
	}

	j.watchComet(c.Discovery)
	return j
}

type kafkaConsumer struct {
	consumer *cluster.Consumer
}

func NewKafka(c *conf.Config) *kafkaConsumer {
	return &kafkaConsumer{
		consumer: newKafkaSub(c.Kafka),
	}
}

func (c *kafkaConsumer) Close() error {
	return c.consumer.Close()
}

type natsConsumer struct {
	consumer *nats.Conn
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

// Close close resounces.
func (j *Job) Close() error {
	if j.consumer != nil {
		return j.consumer.Close()
	}
	return nil
}

func (j *Job) Consume() {
	if j.consumer != nil {
		j.consumer.Consume(j)
	} else {
		log.Errorf("----------> error(%v)", errors.New("consumer is NIL"))
	}

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

func (j *Job) watchComet(c *naming.Config) {
	dis := naming.New(c)
	resolver := dis.Build("goim.comet")
	event := resolver.Watch()
	select {
	case _, ok := <-event:
		if !ok {
			panic("watchComet init failed")
		}
		if ins, ok := resolver.Fetch(); ok {
			if err := j.newAddress(ins); err != nil {
				panic(err)
			}
			log.Infof("watchComet init newAddress:%+v", ins)
		}
	case <-time.After(10 * time.Second):
		log.Error("watchComet init instances timeout")
	}
	go func() {
		for {
			if _, ok := <-event; !ok {
				log.Info("watchComet exit")
				return
			}
			ins, ok := resolver.Fetch()
			if ok {
				if err := j.newAddress(ins); err != nil {
					log.Errorf("watchComet newAddress(%+v) error(%+v)", ins, err)
					continue
				}
				log.Infof("watchComet change newAddress:%+v", ins)
			}
		}
	}()
}

func (j *Job) newAddress(insMap map[string][]*naming.Instance) error {
	ins := insMap[j.c.Env.Zone]
	if len(ins) == 0 {
		return fmt.Errorf("watchComet instance is empty")
	}
	comets := map[string]*Comet{}
	for _, in := range ins {
		if old, ok := j.cometServers[in.Hostname]; ok {
			comets[in.Hostname] = old
			continue
		}
		c, err := NewComet(in, j.c.Comet)
		if err != nil {
			log.Errorf("watchComet NewComet(%+v) error(%v)", in, err)
			return err
		}
		comets[in.Hostname] = c
		log.Infof("watchComet AddComet grpc:%+v", in)
	}
	for key, old := range j.cometServers {
		if _, ok := comets[key]; !ok {
			old.cancel()
			log.Infof("watchComet DelComet:%s", key)
		}
	}
	j.cometServers = comets
	return nil
}

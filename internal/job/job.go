package job

import (
	"errors"
	"sync"
	pb "github.com/Terry-Mao/goim/api/logic/grpc"
	"github.com/Terry-Mao/goim/internal/job/conf"
	"github.com/bilibili/discovery/naming"
	"github.com/gogo/protobuf/proto"
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

// New new a push job.
func New(c *conf.Config) *Job {
	j := &Job{
		c: c,
		// 	consumer: newKafkaSub(c.Kafka),
		rooms: make(map[string]*Room),
	}
	if c.UseNats {
		j.consumer = NewNats(c)
	} else {
		j.consumer = NewKafka(c)
	}

	WatchComet(j, c.Discovery)
	return j
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

	for {
		select {
		case err := <-j.consumer.Errors():
			log.Errorf("consumer error(%v)", err)
		case n := <-j.consumer.Notifications():
			log.Infof("consumer rebalanced(%v)", n)
		case msg, ok := <-j.consumer.Messages():
			if !ok {
				return
			}
			j.consumer.MarkOffset(msg, "")
			// process push message
			pushMsg := new(pb.PushMsg)
			if err := proto.Unmarshal(msg.Value, pushMsg); err != nil {
				log.Errorf("proto.Unmarshal(%v) error(%v)", msg, err)
				continue
			}
			if err := j.push(context.Background(), pushMsg); err != nil {
				log.Errorf("j.push(%v) error(%v)", pushMsg, err)
			}
			log.Infof("consume: %s/%d/%d\t%s\t%+v", msg.Topic, msg.Partition, msg.Offset, msg.Key, pushMsg)
		}
	}
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
			if err := j.newAddress(ins.Instances); err != nil {
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
				if err := j.newAddress(ins.Instances); err != nil {
					log.Errorf("watchComet newAddress(%+v) error(%+v)", ins, err)
					continue
				}
				log.Infof("watchComet change newAddress:%+v", ins)
			}
		}
	}()
}


}

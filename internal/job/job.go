package job

import (
	"errors"
	"sync"

	"github.com/Terry-Mao/goim/internal/job/conf"

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

}

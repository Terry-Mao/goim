package job

import (
	"fmt"
	"time"

	"github.com/Bilibili/discovery/naming"
	log "github.com/golang/glog"

	"github.com/Terry-Mao/goim/internal/job/conf"
)

func WatchComet(j *Job, cfg *conf.DiscoveryConfig) {
	c := &naming.Config{
		Nodes:  cfg.Nodes,
		Region: cfg.Region,
		Zone:   cfg.Zone,
		Env:    cfg.Env,
		Host:   cfg.Host,
	}

	dis := naming.New(c)
	resolver := dis.Build("goim.comet")
	event := resolver.Watch()
	select {
	case _, ok := <-event:
		if !ok {
			panic("WatchComet init failed")
		}
		if ins, ok := resolver.Fetch(); ok {
			if err := newAddress(j, ins); err != nil {
				panic(err)
			}
			log.Infof("WatchComet init newAddress:%+v", ins)
		}
	case <-time.After(10 * time.Second):
		log.Error("WatchComet init instances timeout")
	}
	go func() {
		for {
			if _, ok := <-event; !ok {
				log.Info("WatchComet exit")
				return
			}
			ins, ok := resolver.Fetch()
			if ok {
				if err := newAddress(j, ins); err != nil {
					log.Errorf("WatchComet newAddress(%+v) error(%+v)", ins, err)
					continue
				}
				log.Infof("WatchComet change newAddress:%+v", ins)
			}
		}
	}()
}

func newAddress(j *Job, insMap map[string][]*naming.Instance) error {
	ins := insMap[j.c.Env.Zone]
	if len(ins) == 0 {
		return fmt.Errorf("WatchComet instance is empty")
	}
	comets := map[string]*Comet{}
	for _, in := range ins {
		if old, ok := j.cometServers[in.Hostname]; ok {
			comets[in.Hostname] = old
			continue
		}
		c, err := NewComet(in, j.c.Comet)
		if err != nil {
			log.Errorf("WatchComet NewComet(%+v) error(%v)", in, err)
			return err
		}
		comets[in.Hostname] = c
		log.Infof("WatchComet AddComet grpc:%+v", in)
	}
	for key, old := range j.cometServers {
		if _, ok := comets[key]; !ok {
			old.cancel()
			log.Infof("WatchComet DelComet:%s", key)
		}
	}
	j.cometServers = comets
	return nil
}

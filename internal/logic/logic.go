package logic

import (
	"context"
	"strconv"
	"time"

	"github.com/Bilibili/discovery/naming"
	"github.com/Terry-Mao/goim/internal/logic/conf"
	"github.com/Terry-Mao/goim/internal/logic/dao"
	"github.com/Terry-Mao/goim/internal/logic/model"
	log "github.com/golang/glog"
)

const (
	_onlineTick     = time.Second * 10
	_onlineDeadline = time.Minute * 5
)

// Logic struct
type Logic struct {
	c   *conf.Config
	dis *naming.Discovery
	dao *dao.Dao
	// online
	totalIPs   int64
	totalConns int64
	roomCount  map[string]int32
	// load balancer
	servers      []*naming.Instance
	loadBalancer *LoadBalancer
	regions      map[string]string // province -> region
}

// New init
func New(c *conf.Config) (s *Logic) {
	s = &Logic{
		c:            c,
		dao:          dao.New(c),
		dis:          naming.New(c.Discovery),
		loadBalancer: NewLoadBalancer(),
		regions:      make(map[string]string),
	}
	s.initRegions()
	s.initServer()
	s.loadOnline()
	go s.onlineproc()
	return s
}

// Ping ping resources is ok.
func (s *Logic) Ping(c context.Context) (err error) {
	return s.dao.Ping(c)
}

// Close close resources.
func (s *Logic) Close() {
	s.dao.Close()
}

func (s *Logic) initRegions() {
	for region, ps := range s.c.Regions {
		for _, province := range ps {
			s.regions[province] = region
		}
	}
}

func (s *Logic) initServer() {
	res := s.dis.Build("push.interface.broadcast")
	event := res.Watch()
	select {
	case _, ok := <-event:
		if ok {
			s.newServers(res)
		} else {
			panic("discovery watch failed")
		}
	case <-time.After(10 * time.Second):
		log.Error("discovery start timeout")
	}
	go func() {
		for {
			if _, ok := <-event; !ok {
				return
			}
			s.newServers(res)
		}
	}()
}

func (s *Logic) newServers(res naming.Resolver) {
	if zoneIns, ok := res.Fetch(); ok {
		var (
			totalConns int64
			totalIPs   int64
			ins        []*naming.Instance
		)
		for _, zins := range zoneIns {
			for _, in := range zins {
				if in.Metadata == nil {
					log.Errorf("instance metadata is empty(%+v)", in)
					continue
				}
				if in.Metadata["offline"] == "true" {
					continue
				}
				conns, err := strconv.ParseInt(in.Metadata["conns"], 10, 32)
				if err != nil {
					log.Errorf("strconv.ParseInt(conns:%d) error(%v)", conns, err)
					continue
				}
				ips, err := strconv.ParseInt(in.Metadata["ips"], 10, 32)
				if err != nil {
					log.Errorf("strconv.ParseInt(ips:%d) error(%v)", ips, err)
					continue
				}
				totalConns += conns
				totalIPs += ips
				ins = append(ins, in)
			}
		}
		s.totalConns = totalConns
		s.totalIPs = totalIPs
		s.servers = ins
		s.loadBalancer.Update(ins)
	}
}

func (s *Logic) onlineproc() {
	for {
		time.Sleep(_onlineTick)
		if err := s.loadOnline(); err != nil {
			log.Errorf("onlineproc error(%v)", err)
		}
	}
}

func (s *Logic) loadOnline() (err error) {
	var (
		roomCount = make(map[string]int32)
	)
	for _, server := range s.servers {
		var online *model.Online
		online, err = s.dao.ServerOnline(context.Background(), server.Hostname)
		if err != nil {
			return
		}
		if time.Since(time.Unix(online.Updated, 0)) > _onlineDeadline {
			s.dao.DelServerOnline(context.Background(), server.Hostname)
			continue
		}
		for roomID, count := range online.RoomCount {
			roomCount[roomID] += count
		}
	}
	s.roomCount = roomCount
	return
}

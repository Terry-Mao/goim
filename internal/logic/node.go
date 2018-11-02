package service

import (
	"context"
	"time"

	"github.com/Bilibili/discovery/naming"
	pb "github.com/Terry-Mao/goim/api/logic/grpc"
	log "github.com/golang/glog"
)

// NodeInfos get servers info.
func (s *Service) NodeInfos(c context.Context) (res []*naming.Instance, err error) {
	return s.servers, nil
}

// NodeList get node list.
func (s *Service) NodeList(c context.Context, platform, clientIP string) *pb.NodeListReply {
	reply := &pb.NodeListReply{
		Domain:       s.c.Node.DefaultDomain,
		TcpPort:      int32(s.c.Node.TCPPort),
		WsPort:       int32(s.c.Node.WSPort),
		WssPort:      int32(s.c.Node.WSSPort),
		Heartbeat:    int32(time.Duration(s.c.Node.Heartbeat) / time.Second),
		HeartbeatMax: int32(s.c.Node.HeartbeatMax),
		Backoff: &pb.Backoff{
			MaxDelay:  s.c.Backoff.MaxDelay,
			BaseDelay: s.c.Backoff.BaseDelay,
			Factor:    s.c.Backoff.Factor,
			Jitter:    s.c.Backoff.Jitter,
		},
	}
	domains, addrs := s.nodeAddrs(c, clientIP)
	if platform == "web" {
		reply.Nodes = domains
	} else {
		reply.Nodes = addrs
	}
	if len(reply.Nodes) == 0 {
		reply.Nodes = []string{s.c.Node.DefaultDomain}
	}
	return reply
}

// NodeWeighted debug node weighted.
func (s *Service) NodeWeighted(c context.Context, clientIP string) (interface{}, string, string, error) {
	var (
		region string
	)
	province, err := s.location(c, clientIP)
	if err == nil {
		region = s.regions[province]
	}
	return s.loadBalancer.NodeDetails(region, s.c.Node.RegionWeight), region, province, nil
}

func (s *Service) nodeAddrs(c context.Context, clientIP string) (domains, addrs []string) {
	var (
		region string
	)
	province, err := s.location(c, clientIP)
	if err == nil {
		region = s.regions[province]
	}
	log.Infof("nodeAddrs clientIP:%s region:%s province:%s domains:%v addrs:%v", clientIP, region, province, domains, addrs)
	return s.loadBalancer.NodeAddrs(region, s.c.Node.HostDomain, s.c.Node.RegionWeight)
}

func (s *Service) location(c context.Context, clientIP string) (province string, err error) {
	// TODO find a geolocation of an IP address including province, region and country.
	// province: 上海
	return
}

package service

import (
	"context"
	"time"

	"github.com/Bilibili/discovery/naming"
	pb "github.com/Terry-Mao/goim/api/logic/grpc"
	log "github.com/golang/glog"
)

// ServerInfos get servers info.
func (s *Service) ServerInfos(c context.Context) (res []*naming.Instance, err error) {
	return s.servers, nil
}

// ServerList get server list.
func (s *Service) ServerList(c context.Context, platform, ip string) *pb.ServerListReply {
	reply := &pb.ServerListReply{
		Domain:       s.c.Server.Domain,
		TcpPort:      int32(s.c.Server.TCPPort),
		WsPort:       int32(s.c.Server.WSPort),
		WssPort:      int32(s.c.Server.WSSPort),
		Heartbeat:    int32(time.Duration(s.c.Server.Heartbeat) / time.Second),
		HeartbeatMax: int32(s.c.Server.HeartbeatMax),
		Backoff: &pb.Backoff{
			MaxDelay:  s.c.Backoff.MaxDelay,
			BaseDelay: s.c.Backoff.BaseDelay,
			Factor:    s.c.Backoff.Factor,
			Jitter:    s.c.Backoff.Jitter,
		},
	}
	domains, addrs := s.nodeAddrs(c, ip)
	if platform == "web" {
		reply.Nodes = domains
	} else {
		reply.Nodes = addrs
	}
	if len(reply.Nodes) == 0 {
		reply.Nodes = []string{s.c.Server.Domain}
	}
	return reply
}

// ServerWeight server node details.
func (s *Service) ServerWeight(c context.Context, clientIP string) (interface{}, string, string, error) {
	var (
		region   string
		province string
	)
	if clientIP != "" {
		// TODO	region/province
	}
	return s.loadBalancer.NodeDetails(region, s.c.Server.RegionWeight), region, province, nil
}

func (s *Service) nodeAddrs(c context.Context, ip string) (domains, addrs []string) {
	var region string
	province, err := s.location(c, ip)
	if err == nil {
		region = s.regions[province]
	}
	domains, addrs = s.loadBalancer.NodeAddrs(region, s.c.Server.HostDomain, s.c.Server.RegionWeight)
	log.Infof("nodeAddrs clientIP:%s region:%s province:%s domains:%v addrs:%v", ip, region, province, domains, addrs)
	return
}

func (s *Service) location(c context.Context, ip string) (province string, err error) {
	// TODO find a geolocation of an IP address including province, region and country.
	// province: 中国上海
	return
}

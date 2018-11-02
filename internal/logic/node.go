package logic

import (
	"context"
	"time"

	"github.com/Bilibili/discovery/naming"
	pb "github.com/Terry-Mao/goim/api/logic/grpc"
	"github.com/Terry-Mao/goim/internal/logic/model"
	log "github.com/golang/glog"
)

// NodesInfos get servers info.
func (l *Logic) NodesInfos(c context.Context) (res []*naming.Instance, err error) {
	return l.nodes, nil
}

// NodesWeighted get node list.
func (l *Logic) NodesWeighted(c context.Context, platform, clientIP string) *pb.NodesReply {
	reply := &pb.NodesReply{
		Domain:       l.c.Node.DefaultDomain,
		TcpPort:      int32(l.c.Node.TCPPort),
		WsPort:       int32(l.c.Node.WSPort),
		WssPort:      int32(l.c.Node.WSSPort),
		Heartbeat:    int32(time.Duration(l.c.Node.Heartbeat) / time.Second),
		HeartbeatMax: int32(l.c.Node.HeartbeatMax),
		Backoff: &pb.Backoff{
			MaxDelay:  l.c.Backoff.MaxDelay,
			BaseDelay: l.c.Backoff.BaseDelay,
			Factor:    l.c.Backoff.Factor,
			Jitter:    l.c.Backoff.Jitter,
		},
	}
	domains, addrs := l.nodeAddrs(c, clientIP)
	if platform == model.PlatformWeb {
		reply.Nodes = domains
	} else {
		reply.Nodes = addrs
	}
	if len(reply.Nodes) == 0 {
		reply.Nodes = []string{l.c.Node.DefaultDomain}
	}
	return reply
}

// NodesDebug debug node weighted.
func (l *Logic) NodesDebug(c context.Context, clientIP string) (interface{}, string, string, error) {
	var (
		region string
	)
	province, err := l.location(c, clientIP)
	if err == nil {
		region = l.regions[province]
	}
	return l.loadBalancer.NodeDetails(region, l.c.Node.RegionWeight), region, province, nil
}

func (l *Logic) nodeAddrs(c context.Context, clientIP string) (domains, addrs []string) {
	var (
		region string
	)
	province, err := l.location(c, clientIP)
	if err == nil {
		region = l.regions[province]
	}
	log.Infof("nodeAddrs clientIP:%s region:%s province:%s domains:%v addrs:%v", clientIP, region, province, domains, addrs)
	return l.loadBalancer.NodeAddrs(region, l.c.Node.HostDomain, l.c.Node.RegionWeight)
}

func (l *Logic) location(c context.Context, clientIP string) (province string, err error) {
	// TODO find a geolocation of an IP address including province, region and country.
	// province: 上海
	return
}

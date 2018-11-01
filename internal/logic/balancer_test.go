package logic

import (
	"sort"
	"testing"

	"github.com/Bilibili/discovery/naming"
)

func TestWeightedNode(t *testing.T) {
	nodes := []*weightedNode{
		&weightedNode{fixedWeight: 1, currentWeight: 1, currentConns: 1000000},
		&weightedNode{fixedWeight: 2, currentWeight: 1, currentConns: 1000000},
		&weightedNode{fixedWeight: 3, currentWeight: 1, currentConns: 1000000},
	}
	for i := 0; i < 100; i++ {
		for _, n := range nodes {
			n.calcuateWeight(6, nodes[0].currentConns+nodes[1].currentConns+nodes[2].currentConns, 1.0)
		}
		sort.Slice(nodes, func(i, j int) bool {
			return nodes[i].currentWeight > nodes[j].currentWeight
		})
		nodes[0].chosen()
	}
	ft := float64(nodes[0].fixedWeight + nodes[1].fixedWeight + nodes[2].fixedWeight)
	ct := float64(nodes[0].currentConns + nodes[1].currentConns + nodes[2].currentConns)
	for _, n := range nodes {
		t.Logf("match ratio %d:%d", int(float64(n.fixedWeight)/ft*100*0.6), int(float64(n.currentConns)/ct*100))
	}
}

func TestLoadBalancer(t *testing.T) {
	ss := []*naming.Instance{
		&naming.Instance{
			Region:   "bj",
			Hostname: "01",
			Metadata: map[string]string{
				"weight":   "10",
				"conns":    "240590",
				"ips":      "10",
				"ip_addrs": "ip_bj",
			},
		},
		&naming.Instance{
			Region:   "sh",
			Hostname: "02",
			Metadata: map[string]string{
				"weight":   "10",
				"conns":    "375420",
				"ips":      "10",
				"ip_addrs": "ip_sh",
			},
		},
		&naming.Instance{
			Region:   "gz",
			Hostname: "03",
			Metadata: map[string]string{
				"weight":   "10",
				"conns":    "293430",
				"ips":      "10",
				"ip_addrs": "ip_gz",
			},
		},
	}
	lb := NewLoadBalancer()
	lb.Update(ss)
	for i := 0; i < 5; i++ {
		t.Log(lb.NodeDetails("sh", 1.6))
	}
	for i := 0; i < 5; i++ {
		t.Log(lb.NodeAddrs("sh", ".test", 1.6))
	}
}

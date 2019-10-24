package logic

import (
	"sort"
	"testing"

	"github.com/bilibili/discovery/naming"
	"github.com/Terry-Mao/goim/internal/logic/model"
)

func TestWeightedNode(t *testing.T) {
	nodes := []*weightedNode{
		&weightedNode{fixedWeight: 1, currentWeight: 1, currentConns: 1000000},
		&weightedNode{fixedWeight: 2, currentWeight: 1, currentConns: 1000000},
		&weightedNode{fixedWeight: 3, currentWeight: 1, currentConns: 1000000},
	}
	for i := 0; i < 100; i++ {
		for _, n := range nodes {
			n.calculateWeight(6, nodes[0].currentConns+nodes[1].currentConns+nodes[2].currentConns, 1.0)
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
				model.MetaWeight:    "10",
				model.MetaConnCount: "240590",
				model.MetaIPCount:   "10",
				model.MetaAddrs:     "ip_bj",
			},
		},
		&naming.Instance{
			Region:   "sh",
			Hostname: "02",
			Metadata: map[string]string{
				model.MetaWeight:    "10",
				model.MetaConnCount: "375420",
				model.MetaIPCount:   "10",
				model.MetaAddrs:     "ip_sh",
			},
		},
		&naming.Instance{
			Region:   "gz",
			Hostname: "03",
			Metadata: map[string]string{
				model.MetaWeight:    "10",
				model.MetaConnCount: "293430",
				model.MetaIPCount:   "10",
				model.MetaAddrs:     "ip_gz",
			},
		},
	}
	lb := NewLoadBalancer()
	lb.Update(ss)
	for i := 0; i < 5; i++ {
		t.Log(lb.NodeAddrs("sh", ".test", 1.6))
	}
}

package logic

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/bilibili/discovery/naming"
	"github.com/Terry-Mao/goim/internal/logic/model"
	log "github.com/golang/glog"
)

const (
	_minWeight = 1
	_maxWeight = 1 << 20
	_maxNodes  = 5
)

type weightedNode struct {
	region        string
	hostname      string
	addrs         []string
	fixedWeight   int64
	currentWeight int64
	currentConns  int64
	updated       int64
}

func (w *weightedNode) String() string {
	return fmt.Sprintf("region:%s fixedWeight:%d, currentWeight:%d, currentConns:%d", w.region, w.fixedWeight, w.currentWeight, w.currentConns)
}

func (w *weightedNode) chosen() {
	w.currentConns++
}

func (w *weightedNode) reset() {
	w.currentWeight = 0
}

func (w *weightedNode) calculateWeight(totalWeight, totalConns int64, gainWeight float64) {
	fixedWeight := float64(w.fixedWeight) * gainWeight
	totalWeight += int64(fixedWeight) - w.fixedWeight
	if totalConns > 0 {
		weightRatio := fixedWeight / float64(totalWeight)
		var connRatio float64
		if totalConns != 0 {
			connRatio = float64(w.currentConns) / float64(totalConns) * 0.5
		}
		diff := weightRatio - connRatio
		multiple := diff * float64(totalConns)
		floor := math.Floor(multiple)
		if floor-multiple >= -0.5 {
			w.currentWeight = int64(fixedWeight + floor)
		} else {
			w.currentWeight = int64(fixedWeight + math.Ceil(multiple))
		}
		if diff < 0 {
			// we always return the max from minWeight and calculated Current weight
			if _minWeight > w.currentWeight {
				w.currentWeight = _minWeight
			}
		} else {
			// we always return the min from maxWeight and calculated Current weight
			if _maxWeight < w.currentWeight {
				w.currentWeight = _maxWeight
			}
		}
	} else {
		w.reset()
	}
}

// LoadBalancer load balancer.
type LoadBalancer struct {
	totalConns  int64
	totalWeight int64
	nodes       map[string]*weightedNode
	nodesMutex  sync.Mutex
}

// NewLoadBalancer new a load balancer.
func NewLoadBalancer() *LoadBalancer {
	lb := &LoadBalancer{
		nodes: make(map[string]*weightedNode),
	}
	return lb
}

// Size return node size.
func (lb *LoadBalancer) Size() int {
	return len(lb.nodes)
}

func (lb *LoadBalancer) weightedNodes(region string, regionWeight float64) (nodes []*weightedNode) {
	for _, n := range lb.nodes {
		var gainWeight = float64(1.0)
		if n.region == region {
			gainWeight *= regionWeight
		}
		n.calculateWeight(lb.totalWeight, lb.totalConns, gainWeight)
		nodes = append(nodes, n)
	}
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].currentWeight > nodes[j].currentWeight
	})
	if len(nodes) > 0 {
		nodes[0].chosen()
		lb.totalConns++
	}
	return
}

// NodeAddrs return node addrs.
func (lb *LoadBalancer) NodeAddrs(region, domain string, regionWeight float64) (domains, addrs []string) {
	lb.nodesMutex.Lock()
	nodes := lb.weightedNodes(region, regionWeight)
	lb.nodesMutex.Unlock()
	for i, n := range nodes {
		if i == _maxNodes {
			break
		}
		domains = append(domains, n.hostname+domain)
		addrs = append(addrs, n.addrs...)
	}
	return
}

// Update update server nodes.
func (lb *LoadBalancer) Update(ins []*naming.Instance) {
	var (
		totalConns  int64
		totalWeight int64
		nodes       = make(map[string]*weightedNode, len(ins))
	)
	if len(ins) == 0 || float32(len(ins))/float32(len(lb.nodes)) < 0.5 {
		log.Errorf("load balancer update src:%d target:%d less than half", len(lb.nodes), len(ins))
		return
	}
	lb.nodesMutex.Lock()
	for _, in := range ins {
		if old, ok := lb.nodes[in.Hostname]; ok && old.updated == in.LastTs {
			nodes[in.Hostname] = old
			totalConns += old.currentConns
			totalWeight += old.fixedWeight
		} else {
			meta := in.Metadata
			weight, err := strconv.ParseInt(meta[model.MetaWeight], 10, 32)
			if err != nil {
				log.Errorf("instance(%+v) strconv.ParseInt(weight:%s) error(%v)", in, meta[model.MetaWeight], err)
				continue
			}
			conns, err := strconv.ParseInt(meta[model.MetaConnCount], 10, 32)
			if err != nil {
				log.Errorf("instance(%+v) strconv.ParseInt(conns:%s) error(%v)", in, meta[model.MetaConnCount], err)
				continue
			}
			nodes[in.Hostname] = &weightedNode{
				region:       in.Region,
				hostname:     in.Hostname,
				fixedWeight:  weight,
				currentConns: conns,
				addrs:        strings.Split(meta[model.MetaAddrs], ","),
				updated:      in.LastTs,
			}
			totalConns += conns
			totalWeight += weight
		}
	}
	lb.nodes = nodes
	lb.totalConns = totalConns
	lb.totalWeight = totalWeight
	lb.nodesMutex.Unlock()
}

package resolver

import (
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/Bilibili/discovery/naming"
	log "github.com/golang/glog"
	"github.com/pkg/errors"
	"google.golang.org/grpc/resolver"
)

const (
	// Scheme is the scheme of discovery address
	Scheme = "grpc"
)

var (
	_ resolver.Resolver = &Resolver{}
	_ resolver.Builder  = &Builder{}

	zone = os.Getenv("ZONE")
)

// Register register resolver builder if nil.
func Register(b naming.Builder) {
	resolver.Register(&Builder{b})
}

// Builder is also a resolver builder.
// It's build() function always returns itself.
type Builder struct {
	naming.Builder
}

// Build returns itself for Resolver, because it's both a builder and a resolver.
func (b *Builder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOption) (resolver.Resolver, error) {
	clusters := map[string]struct{}{}
	str := strings.SplitN(target.Endpoint, "?", 2)
	if len(str) == 0 {
		return nil, errors.Errorf("grpc resolver: parse target.Endpoint(%s) failed! the endpoint is empty", target.Endpoint)
	} else if len(str) == 2 {
		m, err := url.ParseQuery(str[1])
		if err == nil {
			for _, c := range m["cluster"] {
				clusters[c] = struct{}{}
			}
			zones := m["zone"]
			if len(zones) > 0 {
				zone = zones[0]
			}
		}
	}
	r := &Resolver{
		nr:       b.Builder.Build(str[0]),
		cc:       cc,
		quit:     make(chan struct{}, 1),
		clusters: clusters,
		zone:     zone,
	}
	go r.updateproc()
	return r, nil
}

// Resolver watches for the updates on the specified target.
// Updates include address updates and service config updates.
type Resolver struct {
	nr   naming.Resolver
	cc   resolver.ClientConn
	quit chan struct{}

	clusters map[string]struct{}
	zone     string
}

// Close is a noop for Resolver.
func (r *Resolver) Close() {
	select {
	case r.quit <- struct{}{}:
		r.nr.Close()
	default:
	}
}

// ResolveNow is a noop for Resolver.
func (r *Resolver) ResolveNow(o resolver.ResolveNowOption) {
}

func (r *Resolver) updateproc() {
	event := r.nr.Watch()
	for {
		select {
		case <-r.quit:
			return
		case _, ok := <-event:
			if !ok {
				return
			}
		}
		if insMap, ok := r.nr.Fetch(); ok {
			instances, ok := insMap[r.zone]
			if !ok {
				for _, value := range insMap {
					instances = append(instances, value...)
				}
			}
			r.newAddress(instances)
		}
	}
}

func (r *Resolver) newAddress(instances []*naming.Instance) {
	if len(instances) <= 0 {
		return
	}
	addrs := make([]resolver.Address, 0, len(instances))
	for _, ins := range instances {
		if len(r.clusters) > 0 {
			if _, ok := r.clusters[ins.Metadata["cluster"]]; !ok {
				continue
			}
		}
		var weight int64
		if weight, _ = strconv.ParseInt(ins.Metadata["weight"], 10, 64); weight <= 0 {
			weight = 10
		}
		var rpc string
		for _, a := range ins.Addrs {
			u, err := url.Parse(a)
			if err == nil && u.Scheme == Scheme {
				rpc = u.Host
			}
		}
		if rpc == "" {
			log.Warningf("grpc resolver: invalid rpc address(%s,%s,%v) found!", ins.AppID, ins.Hostname, ins.Addrs)
			continue
		}
		addr := resolver.Address{
			Addr:       rpc,
			Type:       resolver.Backend,
			ServerName: ins.AppID,
			Metadata:   ins.Metadata,
		}
		addrs = append(addrs, addr)
	}
	r.cc.NewAddress(addrs)
}

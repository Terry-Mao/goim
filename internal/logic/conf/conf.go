package conf

import (
	"flag"
	"os"
	"strconv"
	"time"

	"github.com/Bilibili/discovery/naming"
	xtime "github.com/Terry-Mao/goim/pkg/time"

	"github.com/BurntSushi/toml"
)

var (
	confPath  string
	region    string
	zone      string
	deployEnv string
	host      string
	weight    int64
	// Conf config
	Conf = &Config{}
)

func init() {
	var (
		defHost, _   = os.Hostname()
		defWeight, _ = strconv.ParseInt(os.Getenv("WEIGHT"), 10, 32)
	)
	flag.StringVar(&confPath, "conf", "logic.toml", "default config path")
	flag.StringVar(&region, "region", os.Getenv("REGION"), "distribution region, likes bj/sh/gz/hk/jp/sv/de")
	flag.StringVar(&zone, "zone", os.Getenv("ZONE"), "deployment zone, likes sh001/sh002/sh003")
	flag.StringVar(&deployEnv, "deploy.env", os.Getenv("DEPLOY_ENV"), "deployment environment, likes dev/fat/uat/pre/prod")
	flag.StringVar(&host, "host", defHost, "unique hostname")
	flag.Int64Var(&weight, "weight", defWeight, "server weight for load balancer")
}

// Config .
type Config struct {
	Env        *Env
	Discovery  *naming.Config
	RPCClient  *RPCClient
	RPCServer  *RPCServer
	HTTPServer *HTTPServer
	Kafka      *Kafka
	Redis      *Redis
	Node       *Node
	Backoff    *Backoff
	Regions    map[string][]string
}

// Env is env config.
type Env struct {
	Region    string
	Zone      string
	DeployEnv string
	Host      string
	Weight    int64
}

func (e *Env) fix() {
	if e.Region == "" {
		e.Region = region
	}
	if e.Zone == "" {
		e.Zone = zone
	}
	if e.DeployEnv == "" {
		e.DeployEnv = deployEnv
	}
	if e.Host == "" {
		e.Host = host
	}
	if e.Weight <= 0 {
		e.Weight = weight
	}
}

// Node .
type Node struct {
	DefaultDomain string
	HostDomain    string
	TCPPort       int
	WSPort        int
	WSSPort       int
	HeartbeatMax  int
	Heartbeat     xtime.Duration
	RegionWeight  float64
}

// Backoff .
type Backoff struct {
	MaxDelay  int32
	BaseDelay int32
	Factor    float32
	Jitter    float32
}

// Redis .
type Redis struct {
	Network      string
	Addr         string
	Auth         string
	Active       int
	Idle         int
	DialTimeout  xtime.Duration
	ReadTimeout  xtime.Duration
	WriteTimeout xtime.Duration
	IdleTimeout  xtime.Duration
	Expire       xtime.Duration
}

// Kafka .
type Kafka struct {
	Topic   string
	Brokers []string
}

// RPCClient is RPC client config.
type RPCClient struct {
	Dial    xtime.Duration
	Timeout xtime.Duration
}

func (r *RPCClient) fix() {
	if r.Dial <= 0 {
		r.Dial = xtime.Duration(time.Second)
	}
	if r.Timeout <= 0 {
		r.Timeout = xtime.Duration(time.Second)
	}
}

// RPCServer is RPC server config.
type RPCServer struct {
	Network           string
	Addr              string
	Timeout           xtime.Duration
	IdleTimeout       xtime.Duration
	MaxLifeTime       xtime.Duration
	ForceCloseWait    xtime.Duration
	KeepAliveInterval xtime.Duration
	KeepAliveTimeout  xtime.Duration
}

func (r *RPCServer) fix() {
	if r.Network == "" {
		r.Network = "tcp"
	}
	if r.Timeout <= 0 {
		r.Timeout = xtime.Duration(time.Second)
	}
	if r.IdleTimeout <= 0 {
		r.IdleTimeout = xtime.Duration(time.Second * 60)
	}
	if r.MaxLifeTime <= 0 {
		r.MaxLifeTime = xtime.Duration(time.Hour * 2)
	}
	if r.ForceCloseWait <= 0 {
		r.ForceCloseWait = xtime.Duration(time.Second * 20)
	}
	if r.KeepAliveInterval <= 0 {
		r.KeepAliveInterval = xtime.Duration(time.Second * 60)
	}
	if r.KeepAliveTimeout <= 0 {
		r.KeepAliveTimeout = xtime.Duration(time.Second * 20)
	}
}

// HTTPServer is http server config.
type HTTPServer struct {
	Network      string
	Addr         string
	ReadTimeout  xtime.Duration
	WriteTimeout xtime.Duration
}

// Init init conf
func Init() error {
	return local()
}

func local() (err error) {
	if _, err = toml.DecodeFile(confPath, &Conf); err != nil {
		return
	}
	Conf.fix()
	return
}
func (c *Config) fix() {
	if c.Env == nil {
		c.Env = new(Env)
	}
	c.Env.fix()
	if c.RPCClient != nil {
		c.RPCClient.fix()
	}
	if c.RPCServer != nil {
		c.RPCServer.fix()
	}
	if c.Discovery.Region == "" {
		c.Discovery.Region = c.Env.Region
	}
	if c.Discovery.Zone == "" {
		c.Discovery.Zone = c.Env.Zone
	}
	if c.Discovery.Env == "" {
		c.Discovery.Env = c.Env.DeployEnv
	}
	if c.Discovery.Host == "" {
		c.Discovery.Host = c.Env.Host
	}
}

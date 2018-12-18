package conf

import (
	"flag"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Bilibili/discovery/naming"
	"github.com/BurntSushi/toml"
	xtime "github.com/Terry-Mao/goim/pkg/time"
)

var (
	confPath  string
	region    string
	zone      string
	deployEnv string
	host      string
	offline   bool
	weight    int64
	addrs     string
	// Conf config
	Conf = &Config{}
)

func init() {
	var (
		defHost, _    = os.Hostname()
		defOffline, _ = strconv.ParseBool(os.Getenv("OFFLINE"))
		defWeight, _  = strconv.ParseInt(os.Getenv("WEIGHT"), 10, 32)
		defAddrs      = os.Getenv("ADDRS")
	)
	flag.StringVar(&confPath, "conf", "comet.toml", "default config path")
	flag.StringVar(&region, "region", os.Getenv("REGION"), "distribution region, likes bj/sh/gz/hk/jp/sv/de")
	flag.StringVar(&zone, "zone", os.Getenv("ZONE"), "deployment zone, likes sh001/sh002/sh003")
	flag.StringVar(&deployEnv, "deploy.env", os.Getenv("DEPLOY_ENV"), "deployment environment, likes dev/fat/uat/pre/prod")
	flag.StringVar(&host, "host", defHost, "unique hostname")
	flag.BoolVar(&offline, "offline", defOffline, "server offline")
	flag.Int64Var(&weight, "weight", defWeight, "server weight for load balancer")
	flag.StringVar(&addrs, "addrs", defAddrs, "server public ip addrs")
}

// Config is comet config.
type Config struct {
	Debug        bool
	MaxProc      int
	ServerTick   xtime.Duration
	OnlineTick   xtime.Duration
	Discovery    *naming.Config
	TCP          *TCP
	WebSocket    *WebSocket
	Timer        *Timer
	ProtoSection *ProtoSection
	Whitelist    *Whitelist
	Bucket       *Bucket
	RPCClient    *RPCClient
	RPCServer    *RPCServer
	Env          *Env
}

// Env is env config.
type Env struct {
	Region    string
	Zone      string
	DeployEnv string
	Host      string
	Weight    int64
	Offline   bool
	Addrs     []string
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
	if !e.Offline {
		e.Offline = offline
	}
	if len(e.Addrs) == 0 {
		e.Addrs = strings.Split(addrs, ",")
	}
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

// TCP is tcp config.
type TCP struct {
	Bind         []string
	Sndbuf       int
	Rcvbuf       int
	Keepalive    bool
	Reader       int
	ReadBuf      int
	ReadBufSize  int
	Writer       int
	WriteBuf     int
	WriteBufSize int
}

// WebSocket is websocket config.
type WebSocket struct {
	Bind        []string
	TLSOpen     bool
	TLSBind     []string
	CertFile    string
	PrivateFile string
}

// Timer is timer config.
type Timer struct {
	Timer     int
	TimerSize int
}

// ProtoSection is proto section.
type ProtoSection struct {
	HandshakeTimeout xtime.Duration
	WriteTimeout     xtime.Duration
	SvrProto         int
	CliProto         int
}

// Whitelist is white list config.
type Whitelist struct {
	Whitelist []int64
	WhiteLog  string
}

// Bucket is bucket config.
type Bucket struct {
	Size          int
	Channel       int
	Room          int
	RoutineAmount uint64
	RoutineSize   int
}

// Init init conf.
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
	if c.MaxProc <= 0 {
		c.MaxProc = 32
	}
	if c.ServerTick <= 0 {
		c.ServerTick = xtime.Duration(1 * time.Second)
	}
	if c.OnlineTick <= 0 {
		c.OnlineTick = xtime.Duration(10 * time.Second)
	}
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

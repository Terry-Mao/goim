package conf

import (
	"flag"
	"os"
	"strings"
	"time"

	"github.com/Bilibili/discovery/naming"
	"github.com/BurntSushi/toml"
	xtime "github.com/Terry-Mao/goim/pkg/time"
)

var (
	confPath string
	// Conf config
	Conf = &Config{}
)

// Config is comet config.
type Config struct {
	Debug        bool
	MaxProc      int
	ServerTick   xtime.Duration
	OnlineTick   xtime.Duration
	Naming       *naming.Config
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
	Region  string
	Zone    string
	Env     string
	Host    string
	Weight  string
	Offline string
	IPAddrs []string
}

func (e *Env) fix() {
	if e.Region == "" {
		e.Region = os.Getenv("REGION")
	}
	if e.Zone == "" {
		e.Zone = os.Getenv("ZONE")
	}
	if e.Env == "" {
		e.Env = os.Getenv("DEPLOY_ENV")
	}
	if e.Host == "" {
		e.Host, _ = os.Hostname()
	}
	if e.Weight == "" {
		e.Weight = os.Getenv("WEIGHT")
	}
	if e.Offline == "" {
		e.Offline = os.Getenv("OFFLINE")
	}
	if len(e.IPAddrs) == 0 {
		e.IPAddrs = strings.Split(os.Getenv("IP_ADDRS"), ",")
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

func init() {
	flag.StringVar(&confPath, "conf", "", "default config path")
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
	if c.Env == nil {
		c.Env = new(Env)
	}
	c.Env.fix()
}

package conf

import (
	"flag"
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

// Config .
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
}

// TCP config
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

// WebSocket  config
type WebSocket struct {
	Bind        []string
	TLSOpen     bool
	TLSBind     []string
	CertFile    string
	PrivateFile string
}

// Timer config
type Timer struct {
	Timer     int
	TimerSize int
}

// ProtoSection config
type ProtoSection struct {
	HandshakeTimeout xtime.Duration
	WriteTimeout     xtime.Duration
	SvrProto         int
	CliProto         int
}

// Whitelist .
type Whitelist struct {
	Whitelist []int64
	WhiteLog  string
}

// Bucket .
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

// Init init conf
func Init() error {
	return local()
}

func local() (err error) {
	if _, err = toml.DecodeFile(confPath, &Conf); err != nil {
		return
	}
	return
}

// Fix fix config to default.
func (c *Config) Fix() {
	if c.MaxProc <= 0 {
		c.MaxProc = 32
	}
	if c.ServerTick <= 0 {
		c.ServerTick = xtime.Duration(1 * time.Second)
	}
	if c.OnlineTick <= 0 {
		c.OnlineTick = xtime.Duration(10 * time.Second)
	}
}

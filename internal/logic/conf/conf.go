package conf

import (
	"flag"

	"github.com/Bilibili/discovery/naming"
	xtime "github.com/Terry-Mao/goim/pkg/time"

	"github.com/BurntSushi/toml"
)

var (
	confPath string
	// Conf config
	Conf = &Config{}
)

// Config .
type Config struct {
	Discovery *naming.Config
	Kafka     *Kafka
	Redis     *Redis
	Node      *Node
	Backoff   *Backoff
	Regions   map[string][]string
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

func init() {
	flag.StringVar(&confPath, "conf", "", "default config path")
}

// Init init conf
func Init() error {
	return local()
}

func local() (err error) {
	_, err = toml.DecodeFile(confPath, &Conf)
	return
}

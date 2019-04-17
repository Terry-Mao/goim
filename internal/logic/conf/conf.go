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
	Conf *Config
)

func init() {
	var (
		defHost, _   = os.Hostname()
		defWeight, _ = strconv.ParseInt(os.Getenv("WEIGHT"), 10, 32)
	)
	flag.StringVar(&confPath, "conf", "logic-example.toml", "default config path")
	flag.StringVar(&region, "region", os.Getenv("REGION"), "avaliable region. or use REGION env variable, value: sh etc.")
	flag.StringVar(&zone, "zone", os.Getenv("ZONE"), "avaliable zone. or use ZONE env variable, value: sh001/sh002 etc.")
	flag.StringVar(&deployEnv, "deploy.env", os.Getenv("DEPLOY_ENV"), "deploy env. or use DEPLOY_ENV env variable, value: dev/fat1/uat/pre/prod etc.")
	flag.StringVar(&host, "host", defHost, "machine hostname. or use default machine hostname.")
	flag.Int64Var(&weight, "weight", defWeight, "load balancing weight, or use WEIGHT env variable, value: 10 etc.")
}

// Init init config.
func Init() (err error) {
	Conf = Default()
	_, err = toml.DecodeFile(confPath, &Conf)
	return
}

// Default new a config with specified defualt value.
func Default() *Config {
	return &Config{
		Env:       &Env{Region: region, Zone: zone, DeployEnv: deployEnv, Host: host, Weight: weight},
		Discovery: &naming.Config{Region: region, Zone: zone, Env: deployEnv, Host: host},
		HTTPServer: &HTTPServer{
			Network:      "tcp",
			Addr:         "3111",
			ReadTimeout:  xtime.Duration(time.Second),
			WriteTimeout: xtime.Duration(time.Second),
		},
		RPCClient: &RPCClient{Dial: xtime.Duration(time.Second), Timeout: xtime.Duration(time.Second)},
		RPCServer: &RPCServer{
			Network:           "tcp",
			Addr:              "3119",
			Timeout:           xtime.Duration(time.Second),
			IdleTimeout:       xtime.Duration(time.Second * 60),
			MaxLifeTime:       xtime.Duration(time.Hour * 2),
			ForceCloseWait:    xtime.Duration(time.Second * 20),
			KeepAliveInterval: xtime.Duration(time.Second * 60),
			KeepAliveTimeout:  xtime.Duration(time.Second * 20),
		},
		Backoff: &Backoff{MaxDelay: 300, BaseDelay: 3, Factor: 1.8, Jitter: 1.3},
	}
}

// Config config.
type Config struct {
	KafaNatsSwitch bool                `json:"kafaNatsSwitch"`
	Nats           *Nats               `json:"nats"`
	Env            *Env                `json:"env"`
	Discovery      *naming.Config      `json:"discovery"`
	RPCClient      *RPCClient          `json:"rPCClient"`
	RPCServer      *RPCServer          `json:"rPCServer"`
	HTTPServer     *HTTPServer         `json:"hTTPServer"`
	Kafka          *Kafka              `json:"kafka"`
	Redis          *Redis              `json:"redis"`
	Node           *Node               `json:"node"`
	Backoff        *Backoff            `json:"backoff"`
	Regions        map[string][]string `json:"regions"`
}

// Nats configuration for nats
type Nats struct {
	Channel   string `json:"channel"`
	ChannelID string `json:"channelID"`
	Group     string `json:"group"`
	NatsAddr  string `json:"natsAddr"`
	AckInbox  string `json:"ackInbox"`
}

// Env is env config.
type Env struct {
	Region    string `json:"region"`
	Zone      string `json:"zone"`
	DeployEnv string `json:"deployEnv"`
	Host      string `json:"host"`
	Weight    int64  `json:"weight"`
}

// Node node config.
type Node struct {
	DefaultDomain string         `json:"defaultDomain"`
	HostDomain    string         `json:"hostDomain"`
	TCPPort       int            `json:"tCPPort"`
	WSPort        int            `json:"wSPort"`
	WSSPort       int            `json:"wSSPort"`
	HeartbeatMax  int            `json:"heartbeatMax"`
	Heartbeat     xtime.Duration `json:"heartbeat"`
	RegionWeight  float64        `json:"regionWeight"`
}

// Backoff backoff.
type Backoff struct {
	MaxDelay  int32   `json:"maxDelay"`
	BaseDelay int32   `json:"baseDelay"`
	Factor    float32 `json:"factor"`
	Jitter    float32 `json:"jitter"`
}

// Redis .
type Redis struct {
	Network      string         `json:"network"`
	Addr         string         `json:"addr"`
	Auth         string         `json:"auth"`
	Active       int            `json:"active"`
	Idle         int            `json:"idle"`
	DialTimeout  xtime.Duration `json:"dialTimeout"`
	ReadTimeout  xtime.Duration `json:"readTimeout"`
	WriteTimeout xtime.Duration `json:"writeTimeout"`
	IdleTimeout  xtime.Duration `json:"idleTimeout"`
	Expire       xtime.Duration `json:"expire"`
}

// Kafka .
type Kafka struct {
	Topic   string   `json:"topic"`
	Brokers []string `json:"brokers"`
}

// RPCClient is RPC client config.
type RPCClient struct {
	Dial    xtime.Duration `json:"dial"`
	Timeout xtime.Duration `json:"timeout"`
}

// RPCServer is RPC server config.
type RPCServer struct {
	Network           string         `json:"network"`
	Addr              string         `json:"addr"`
	Timeout           xtime.Duration `json:"timeout"`
	IdleTimeout       xtime.Duration `json:"idleTimeout"`
	MaxLifeTime       xtime.Duration `json:"maxLifeTime"`
	ForceCloseWait    xtime.Duration `json:"forceCloseWait"`
	KeepAliveInterval xtime.Duration `json:"keepAliveInterval"`
	KeepAliveTimeout  xtime.Duration `json:"keepAliveTimeout"`
}

// HTTPServer is http server config.
type HTTPServer struct {
	Network      string         `json:"network"`
	Addr         string         `json:"addr"`
	ReadTimeout  xtime.Duration `json:"readTimeout"`
	WriteTimeout xtime.Duration `json:"writeTimeout"`
}

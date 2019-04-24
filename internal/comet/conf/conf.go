package conf

import (
	"flag"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bilibili/discovery/naming"
	"github.com/BurntSushi/toml"
	xtime "github.com/Terry-Mao/goim/pkg/time"
)

var (
	confPath  string
	region    string
	zone      string
	deployEnv string
	host      string
	addrs     string
	weight    int64
	offline   bool
	debug     bool

	// Conf config
	Conf *Config
)

func init() {
	var (
		defHost, _    = os.Hostname()
		defAddrs      = os.Getenv("ADDRS")
		defWeight, _  = strconv.ParseInt(os.Getenv("WEIGHT"), 10, 32)
		defOffline, _ = strconv.ParseBool(os.Getenv("OFFLINE"))
		defDebug, _   = strconv.ParseBool(os.Getenv("DEBUG"))
	)
	flag.StringVar(&confPath, "conf", "comet-example.toml", "default config path.")
	flag.StringVar(&region, "region", os.Getenv("REGION"), "avaliable region. or use REGION env variable, value: sh etc.")
	flag.StringVar(&zone, "zone", os.Getenv("ZONE"), "avaliable zone. or use ZONE env variable, value: sh001/sh002 etc.")
	flag.StringVar(&deployEnv, "deploy.env", os.Getenv("DEPLOY_ENV"), "deploy env. or use DEPLOY_ENV env variable, value: dev/fat1/uat/pre/prod etc.")
	flag.StringVar(&host, "host", defHost, "machine hostname. or use default machine hostname.")
	flag.StringVar(&addrs, "addrs", defAddrs, "server public ip addrs. or use ADDRS env variable, value: 127.0.0.1 etc.")
	flag.Int64Var(&weight, "weight", defWeight, "load balancing weight, or use WEIGHT env variable, value: 10 etc.")
	flag.BoolVar(&offline, "offline", defOffline, "server offline. or use OFFLINE env variable, value: true/false etc.")
	flag.BoolVar(&debug, "debug", defDebug, "server debug. or use DEBUG env variable, value: true/false etc.")
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
		Debug:     debug,
		Env:       &Env{Region: region, Zone: zone, DeployEnv: deployEnv, Host: host, Weight: weight, Addrs: strings.Split(addrs, ","), Offline: offline},
		Discovery: &naming.Config{Region: region, Zone: zone, Env: deployEnv, Host: host},
		RPCClient: &RPCClient{
			Dial:    xtime.Duration(time.Second),
			Timeout: xtime.Duration(time.Second),
		},
		RPCServer: &RPCServer{
			Network:           "tcp",
			Addr:              ":3109",
			Timeout:           xtime.Duration(time.Second),
			IdleTimeout:       xtime.Duration(time.Second * 60),
			MaxLifeTime:       xtime.Duration(time.Hour * 2),
			ForceCloseWait:    xtime.Duration(time.Second * 20),
			KeepAliveInterval: xtime.Duration(time.Second * 60),
			KeepAliveTimeout:  xtime.Duration(time.Second * 20),
		},
		TCP: &TCP{
			Bind:         []string{":3101"},
			Sndbuf:       4096,
			Rcvbuf:       4096,
			KeepAlive:    false,
			Reader:       32,
			ReadBuf:      1024,
			ReadBufSize:  8192,
			Writer:       32,
			WriteBuf:     1024,
			WriteBufSize: 8192,
		},
		Websocket: &Websocket{
			Bind: []string{":3102"},
		},
		Protocol: &Protocol{
			Timer:            32,
			TimerSize:        2048,
			CliProto:         5,
			SvrProto:         10,
			HandshakeTimeout: xtime.Duration(time.Second * 5),
		},
		Bucket: &Bucket{
			Size:          32,
			Channel:       1024,
			Room:          1024,
			RoutineAmount: 32,
			RoutineSize:   1024,
		},
	}
}

// Config is comet config.
type Config struct {
	Debug     bool
	Env       *Env
	Discovery *naming.Config
	TCP       *TCP
	Websocket *Websocket
	Protocol  *Protocol
	Bucket    *Bucket
	RPCClient *RPCClient
	RPCServer *RPCServer
	Whitelist *Whitelist
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

// RPCClient is RPC client config.
type RPCClient struct {
	Dial    xtime.Duration
	Timeout xtime.Duration
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

// TCP is tcp config.
type TCP struct {
	Bind         []string
	Sndbuf       int
	Rcvbuf       int
	KeepAlive    bool
	Reader       int
	ReadBuf      int
	ReadBufSize  int
	Writer       int
	WriteBuf     int
	WriteBufSize int
}

// Websocket is websocket config.
type Websocket struct {
	Bind        []string
	TLSOpen     bool
	TLSBind     []string
	CertFile    string
	PrivateFile string
}

// Protocol is protocol config.
type Protocol struct {
	Timer            int
	TimerSize        int
	SvrProto         int
	CliProto         int
	HandshakeTimeout xtime.Duration
}

// Bucket is bucket config.
type Bucket struct {
	Size          int
	Channel       int
	Room          int
	RoutineAmount uint64
	RoutineSize   int
}

// Whitelist is white list config.
type Whitelist struct {
	Whitelist []int64
	WhiteLog  string
}

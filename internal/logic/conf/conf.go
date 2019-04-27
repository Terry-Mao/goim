package conf

import (
	"flag"
	"os"
	"strconv"
	"time"

	xtime "github.com/Terry-Mao/goim/pkg/time"
	"github.com/bilibili/discovery/naming"

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

// Node node config.
type Node struct {
	DefaultDomain string
	HostDomain    string
	TCPPort       int
	WSPort        int
	WSSPort       int

	// 心跳週期，連線沒有在既定的週期內回應，server就close
	// Heartbeat * HeartbeatMax = 週期時間
	HeartbeatMax int
	Heartbeat    xtime.Duration

	RegionWeight float64
}

// Backoff backoff.
type Backoff struct {
	//
	MaxDelay  int32

	//
	BaseDelay int32

	//
	Factor    float32

	//
	Jitter    float32
}

// Redis
type Redis struct {
	// host
	Network string

	// port
	Addr string

	// password
	Auth string

	// pool內最大連線總數
	Active int

	// 最大保留的閒置連線數
	Idle int

	// 建立連線超時多久後放棄
	DialTimeout xtime.Duration

	// read多久沒回覆則放棄
	ReadTimeout xtime.Duration

	// write多久沒回覆則放棄
	WriteTimeout xtime.Duration

	// 空閒連線多久沒做事就close
	IdleTimeout xtime.Duration

	// redis過期時間
	Expire xtime.Duration
}

// Kafka .
type Kafka struct {
	// Kafka 推送與接收Topic
	Topic   string

	//
	Brokers []string
}

// RPCClient is RPC client config
type RPCClient struct {
	// 沒用到
	Dial xtime.Duration

	// 沒用到
	Timeout xtime.Duration
}

// RPCServer is RPC server config.
type RPCServer struct {
	// host
	Network string

	// port
	Addr string

	// 沒用到
	Timeout xtime.Duration

	// 當連線閒置多久後發送一個`GOAWAY` Framer 封包告知Client說太久沒活動
	//至於Client收到`GOAWAY`後要做什麼目前要自己實現stream，server只是做通知而已，grpc server默認沒開啟此功能
	IdleTimeout xtime.Duration

	// 任何連線只要連線超過某時間就會強制被close，但是在close之前會先發送`GOAWAY`Framer 封包告知Client
	MaxLifeTime xtime.Duration

	// MaxConnectionAge要關閉之前等待的時間
	ForceCloseWait xtime.Duration

	// keepalive頻率(心跳週期)
	KeepAliveInterval xtime.Duration

	// 每次做keepalive完後等待多少秒如果server沒有回應則將此連線close掉
	KeepAliveTimeout xtime.Duration
}

// HTTPServer is http server config.
type HTTPServer struct {
	// host
	Network string

	// port
	Addr string

	// 沒用到
	ReadTimeout xtime.Duration

	// 沒用到
	WriteTimeout xtime.Duration
}

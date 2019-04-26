package conf

import (
	"flag"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	xtime "github.com/Terry-Mao/goim/pkg/time"
	"github.com/bilibili/discovery/naming"
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

// grpc client config
type RPCClient struct {
	// client連線timeout
	Dial xtime.Duration

	// 好像沒用到
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

// TCP is tcp config.
type TCP struct {
	// tcp 要監聽的port
	Bind []string

	// tcp寫資料的緩衝區大小，該緩衝區滿到無法發送時會阻塞，此值通常設定完後系統會自行在多一倍，設定1024會變2304
	Sndbuf int

	// tcp讀取資料的緩衝區大小，該緩衝區為0時會阻塞，此值通常設定完後，系統會自行在多一倍，設定1024會變2304
	Rcvbuf int

	// 是否開啟KeepAlive
	KeepAlive bool

	// 先初始化多少個用於Reader bytes的Pool
	// 每個Pool都會有sync.Mutex，多個pool來分散鎖的競爭
	// 有效提高併發數
	Reader int

	// 每個Reader bytes Pool有多少個Buffer
	ReadBuf int

	// 每個Reader bytes Pool的Buffer能有多大的空間
	ReadBufSize int

	// 先初始化多少個用於Writer bytes的Pool
	// 每個Pool都會有sync.Mutex，多個pool來分散鎖的競爭
	// 有效提高併發數
	Writer int

	// 每個Writer bytes Pool有多少個Buffer
	WriteBuf int

	// 每個Writer bytes Pool的Buffer能有多大的空間
	WriteBufSize int
}

// Websocket is websocket config.
type Websocket struct {
	// Websocket 要監聽的port
	Bind []string

	// 是否打開tls
	TLSOpen bool

	// tls 要監聽的port
	TLSBind []string

	// tls公鑰
	CertFile string

	// tls私鑰
	PrivateFile string
}

// Protocol is protocol config.
type Protocol struct {
	// 先初始化多少個time.Timer
	Timer int

	// 每個time.Timer一開始能接收的TimerData數量
	TimerSize int

	// 每一個連線開grpc接收資料的緩充量，當寫的速度大於讀的速度這時會阻塞，透過調大此值可以有更多緩衝避免阻塞
	SvrProto int

	// 每一個連線開異步Proto結構緩型Pool的大小，跟client透過tcp or websocket傳遞資料做消費速度有關聯
	// 由於寫的速度有可能大於讀的速度，這時會自行close此連線，透過調大此值可以有更多緩衝close
	CliProto int

	// 一開始tcp連線後等待多久沒有請求連至某房間，連線就直接close
	//
	//             -> 送auth資料 ok
	// tcp -> 等待 ->
	//             -> 超時close
	//
	HandshakeTimeout xtime.Duration
}

// Bucket is bucket config.
type Bucket struct {
	// 一開始需要幾個Bucket
	Size int

	// 每個Bucket一開始管理多少個Channel
	Channel int

	// 每個Bucket一開始管理多少個Room
	Room int

	// 每個Bucket開幾個goroutine併發做房間推送
	RoutineAmount uint64

	// 每個房間推送管道最大緩衝量
	RoutineSize int
}

// Whitelist is white list config.
type Whitelist struct {
	Whitelist []int64
	WhiteLog  string
}

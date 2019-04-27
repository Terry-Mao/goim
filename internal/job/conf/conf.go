package conf

import (
	"flag"
	"os"
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
	// Conf config
	Conf *Config
)

func init() {
	var (
		defHost, _ = os.Hostname()
	)
	flag.StringVar(&confPath, "conf", "job-example.toml", "default config path")
	flag.StringVar(&region, "region", os.Getenv("REGION"), "avaliable region. or use REGION env variable, value: sh etc.")
	flag.StringVar(&zone, "zone", os.Getenv("ZONE"), "avaliable zone. or use ZONE env variable, value: sh001/sh002 etc.")
	flag.StringVar(&deployEnv, "deploy.env", os.Getenv("DEPLOY_ENV"), "deploy env. or use DEPLOY_ENV env variable, value: dev/fat1/uat/pre/prod etc.")
	flag.StringVar(&host, "host", defHost, "machine hostname. or use default machine hostname.")
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
		Env:       &Env{Region: region, Zone: zone, DeployEnv: deployEnv, Host: host},
		Discovery: &naming.Config{Region: region, Zone: zone, Env: deployEnv, Host: host},
		Comet:     &Comet{RoutineChan: 1024, RoutineSize: 32},
		Room: &Room{
			Batch:  20,
			Signal: xtime.Duration(time.Second),
			Idle:   xtime.Duration(time.Minute * 15),
		},
	}
}

// Config is job config.
type Config struct {
	Env       *Env
	Kafka     *Kafka
	Discovery *naming.Config
	Comet     *Comet
	Room      *Room
}

// 房間消息聚合
type Room struct {
	// Signal時間內最大緩衝的訊息量，超過就推送給comet
	// 設太小等於Signal沒用，設大太每次都一定要等Signal時間到
	// 默認是20筆
	Batch int

	// 消息聚合等待多久才推送房間消息給comet，默認一秒應該是最好的優化
	// 設小會提高job通知comet的頻率，設太大房間訊息會更延遲
	Signal xtime.Duration

	// 消息聚合goroutine等待多久都沒收到訊息自動close
	Idle xtime.Duration
}

// Comet is comet config.
type Comet struct {
	// 處理訊息推送給comet的chan數量
	RoutineChan int

	// 處理訊息推送給comet的chan的Buffer
	RoutineSize int
}

// Kafka is kafka config.
type Kafka struct {
	Topic   string
	Group   string
	Brokers []string
}

// Env is env config.
type Env struct {
	Region    string
	Zone      string
	DeployEnv string
	Host      string
}

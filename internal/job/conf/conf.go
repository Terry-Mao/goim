package conf

import (
	"flag"
	"os"
	"time"

	"github.com/BurntSushi/toml"

	xtime "github.com/Terry-Mao/goim/pkg/time"
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
		Discovery: &DiscoveryConfig{Region: region, Zone: zone, Env: deployEnv, Host: host},
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
	UseNats   bool             `json:"useNats"`
	Debug     bool             `json:"debug"`
	Env       *Env             `json:"env"`
	Nats      *Nats            `json:"nats"`
	Discovery *DiscoveryConfig `json:"discovery"`
	Kafka     *Kafka           `json:"kafka"`
	Comet     *Comet           `json:"comet"`
	Room      *Room            `json:"room"`
}

// DiscoveryConfig discovery configures.
type DiscoveryConfig struct {
	Nodes  []string
	Region string
	Zone   string
	Env    string
	Host   string
}

// Room is room config.
type Room struct {
	Batch  int            `json:"batch"`
	Signal xtime.Duration `json:"signal"`
	Idle   xtime.Duration `json:"idle"`
}

// Comet is comet config.
type Comet struct {
	RoutineChan int `json:"routineChan"`
	RoutineSize int `json:"routineSize"`
}

// Kafka is kafka config.
type Kafka struct {
	Topic   string   `json:"topic"`
	Group   string   `json:"group"`
	Brokers []string `json:"brokers"`
}

// Nats configuration for nats
type Nats struct {
	Topic    string `json:"topic"`
	TopicID  string `json:"topicID"`
	Brokers  string `json:"brokers"`
	AckInbox string `json:"ackInbox"`
}

// Env is env config.
type Env struct {
	Region    string `json:"region"`
	Zone      string `json:"zone"`
	DeployEnv string `json:"deployEnv"`
	Host      string `json:"host"`
}

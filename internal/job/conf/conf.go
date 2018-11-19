package conf

import (
	"flag"
	"os"
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

func init() {
	flag.StringVar(&confPath, "conf", "job-example.toml", "default config path")
}

// Config is job config.
type Config struct {
	Env       *Env
	Kafka     *Kafka
	Discovery *naming.Config
	Comet     *Comet
	Room      *Room
}

// Room is room config.
type Room struct {
	Refresh xtime.Duration
	Idle    xtime.Duration
	Signal  xtime.Duration
	Batch   int
}

// Comet is comet config.
type Comet struct {
	RoutineChan int
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
	Region string
	Zone   string
	Env    string
	Host   string
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
}

// Init init conf
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
	if c.Discovery.Region == "" {
		c.Discovery.Region = c.Env.Region
	}
	if c.Discovery.Zone == "" {
		c.Discovery.Zone = c.Env.Zone
	}
	if c.Discovery.Env == "" {
		c.Discovery.Env = c.Env.Env
	}
	if c.Discovery.Host == "" {
		c.Discovery.Host = c.Env.Host
	}
	if c.Comet == nil {
		c.Comet = new(Comet)
	}
	if c.Comet.RoutineChan <= 0 {
		c.Comet.RoutineChan = 1024
	}
	if c.Comet.RoutineSize <= 0 {
		c.Comet.RoutineSize = 32
	}
	if c.Room == nil {
		c.Room = new(Room)
	}
	if c.Room.Refresh <= 0 {
		c.Room.Refresh = xtime.Duration(time.Second)
	}
	if c.Room.Idle <= 0 {
		c.Room.Idle = xtime.Duration(time.Minute)
	}
	if c.Room.Signal <= 0 {
		c.Room.Signal = xtime.Duration(time.Second)
	}
	if c.Room.Batch <= 0 {
		c.Room.Batch = 20
	}
}

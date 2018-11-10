package conf

import (
	"flag"
	"os"

	"github.com/Bilibili/discovery/naming"
	"github.com/BurntSushi/toml"
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
}

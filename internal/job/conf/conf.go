package conf

import (
	"os"

	"github.com/BurntSushi/toml"
)

var (
	confPath string
	// Conf config
	Conf = &Config{}
)

// Config is job config.
type Config struct {
	Env *Env
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

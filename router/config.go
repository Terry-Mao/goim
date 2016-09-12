package main

import (
	"flag"
	"runtime"
	"time"

	"github.com/Terry-Mao/goconf"
)

var (
	gconf    *goconf.Config
	Conf     *Config
	confFile string
)

func init() {
	flag.StringVar(&confFile, "c", "./router-example.conf", " set router config file path")
}

type Config struct {
	// base section
	PidFile    string   `goconf:"base:pidfile"`
	Dir        string   `goconf:"base:dir"`
	Log        string   `goconf:"base:log"`
	MaxProc    int      `goconf:"base:maxproc"`
	PprofAddrs []string `goconf:"base:pprof.addrs:,"`
	// rpc
	RPCAddrs []string `goconf:"rpc:addrs:,"`
	// bucket
	Bucket            int           `goconf:"bucket:bucket"`
	Server            int           `goconf:"bucket:server"`
	Cleaner           int           `goconf:"bucket:cleaner"`
	BucketCleanPeriod time.Duration `goconf:"bucket:clean.period:time"`
	// session
	Session       int           `goconf:"session:session"`
	SessionExpire time.Duration `goconf:"session:expire:time"`
	// monitor
	MonitorOpen  bool     `goconf:"monitor:open"`
	MonitorAddrs []string `goconf:"monitor:addrs:,"`
}

func NewConfig() *Config {
	return &Config{
		// base section
		PidFile:    "/tmp/goim-router.pid",
		Dir:        "./",
		Log:        "./router-log.xml",
		MaxProc:    runtime.NumCPU(),
		PprofAddrs: []string{"localhost:6971"},
		// rpc
		RPCAddrs: []string{"localhost:9090"},
		// bucket
		Bucket:            runtime.NumCPU(),
		Server:            5,
		Cleaner:           1000,
		BucketCleanPeriod: time.Hour * 1,
		// session
		Session:       1000,
		SessionExpire: time.Hour * 1,
	}
}

// InitConfig init the global config.
func InitConfig() (err error) {
	Conf = NewConfig()
	gconf = goconf.New()
	if err = gconf.Parse(confFile); err != nil {
		return err
	}
	if err := gconf.Unmarshal(Conf); err != nil {
		return err
	}
	return nil
}

func ReloadConfig() (*Config, error) {
	conf := NewConfig()
	ngconf, err := gconf.Reload()
	if err != nil {
		return nil, err
	}
	if err := ngconf.Unmarshal(conf); err != nil {
		return nil, err
	}
	gconf = ngconf
	return conf, nil
}

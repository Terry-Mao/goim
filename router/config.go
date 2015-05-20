package main

import (
	"flag"
	"github.com/Terry-Mao/goconf"
	"runtime"
	"time"
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
	PidFile   string   `goconf:"base:pidfile"`
	Dir       string   `goconf:"base:dir"`
	Log       string   `goconf:"base:log"`
	MaxProc   int      `goconf:"base:maxproc"`
	PprofBind []string `goconf:"base:pprof.bind:,"`
	StatBind  []string `goconf:"base:stat.bind:,"`
	// http
	HttpBind         []string      `goconf:"http:bind:,"`
	HttpReadTimeout  time.Duration `goconf:"http:read.timeout:time"`
	HttpWriteTimeout time.Duration `goconf:"http:write.timeout:time"`
	// rpc
	RPCBind []string `goconf:"rpc:bind:,"`
	// bucket
	SubBucketNum   int `goconf:"bucket:subbucket.num"`
	TopicBucketNum int `goconf:"bucket:topicbucket.num"`
}

func NewConfig() *Config {
	return &Config{
		// base section
		PidFile:   "/tmp/gopush-cluster-comet.pid",
		Dir:       "./",
		Log:       "./log/xml",
		MaxProc:   runtime.NumCPU(),
		PprofBind: []string{"localhost:6971"},
		StatBind:  []string{"localhost:6972"},
		// http
		HttpBind:         []string{"localhost:8080"},
		HttpReadTimeout:  5 * time.Second,
		HttpWriteTimeout: 5 * time.Second,
		// rpc
		RPCBind: []string{"localhost:9090"},
		// bucket
		SubBucketNum:   1,
		TopicBucketNum: 1,
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

package main

import (
	"flag"
	"github.com/Terry-Mao/goconf"
	"runtime"
)

var (
	gconf    *goconf.Config
	Conf     *Config
	confFile string
)

func init() {
	flag.StringVar(&confFile, "c", "./comet.conf", " set comet config file path")
}

type Config struct {
	// base section
	PidFile   string   `goconf:"base:pidfile"`
	Dir       string   `goconf:"base:dir"`
	Log       string   `goconf:"base:log"`
	MaxProc   int      `goconf:"base:maxproc"`
	PprofBind []string `goconf:"base:pprof.bind:,"`
	// proto section
	TCPAddr string `goconf:"proto:tcp.addr"`
	Sndbuf  int    `goconf:"proto:sndbuf:memory"`
	Rcvbuf  int    `goconf:"proto:rcvbuf:memory"`
	// crypto
	RSAPublic string `goconf:"crypto:rsa.public"`
	// sub
	SubKey string `goconf:sub:sub.key`
}

func NewConfig() *Config {
	return &Config{
		// base section
		PidFile:   "/tmp/gopush-cluster-comet.pid",
		Dir:       "./",
		Log:       "./log/xml",
		MaxProc:   runtime.NumCPU(),
		PprofBind: []string{"localhost:6971"},
		// proto section
		TCPAddr: "localhost:8080",
		Sndbuf:  2048,
		Rcvbuf:  256,
		// crypto
		RSAPublic: "./pub.pem",
		// sub
		SubKey: "Terry-Mao",
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

package main

import (
	"flag"
	"runtime"

	"github.com/Terry-Mao/goconf"
)

var (
	gconf    *goconf.Config
	Conf     *Config
	confFile string
)

func init() {
	flag.StringVar(&confFile, "c", "./client.conf", " set client config file path")
}

type Config struct {
	// base section
	PidFile string `goconf:"base:pidfile"`
	Dir     string `goconf:"base:dir"`
	Log     string `goconf:"base:log"`
	MaxProc int    `goconf:"base:maxproc"`
	// cert
	CertFile string `goconf:"cert:cert.file"`
	// proto section
	TCPAddr       string `goconf:"proto:tcp.addr"`
	WebsocketAddr string `goconf:"proto:websocket.addr"`
	Sndbuf        int    `goconf:"proto:sndbuf:memory"`
	Rcvbuf        int    `goconf:"proto:rcvbuf:memory"`
	Type          int    `goconf:"proto:type"`
	// sub
	SubKey string `goconf:sub:sub.key`
}

func NewConfig() *Config {
	return &Config{
		// base section
		PidFile: "/tmp/goim-client.pid",
		Dir:     "./",
		Log:     "./log.xml",
		MaxProc: runtime.NumCPU(),
		// proto section
		TCPAddr:       "localhost:8080",
		WebsocketAddr: "localhost:8090",
		Sndbuf:        2048,
		Rcvbuf:        256,
		Type:          ProtoTCP,
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

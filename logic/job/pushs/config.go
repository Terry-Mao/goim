package main

import (
	"flag"
	"github.com/Terry-Mao/goconf"
	"strconv"
)

var (
	gconf    *goconf.Config
	Conf     *Config
	confFile string
)

func init() {
	flag.StringVar(&confFile, "c", "./push-job.conf", " set config file path")
}

type Config struct {
	Log               string            `goconf:"base:log"`
	RouterAddrs       []string          `goconf:"router:addr:,"`
	ZKAddrs           []string          `goconf:"kafka:zookeeper.list:,"`
	ZKRoot            string            `goconf:"kafka:zkroot"`
	KafkaTopic        string            `goconf:"kafka:topic"`
	Comets            map[int32]string  `goconf:"-"`
	RouterRPCNetworks []string          `goconf:"router:networks:,"`
	RouterRPCAddrs    map[string]string `-`
}

func NewConfig() *Config {
	return &Config{
		Comets:         make(map[int32]string),
		ZKRoot:         "",
		KafkaTopic:     "kafka_topic_push",
		RouterRPCAddrs: make(map[int32]string),
	}
}

// InitConfig init the global config.
func InitConfig() (err error) {
	Conf = NewConfig()
	gconf = goconf.New()
	if err = gconf.Parse(confFile); err != nil {
		return err
	}
	if err = gconf.Unmarshal(Conf); err != nil {
		return err
	}
	var serverIDi int64
	for _, serverID := range gconf.Get("comets").Keys() {
		addr, err := gconf.Get("comets").String(serverID)
		if err != nil {
			return err
		}
		serverIDi, err = strconv.ParseInt(serverID, 10, 32)
		if err != nil {
			return err
		}

		Conf.Comets[int32(serverIDi)] = addr
	}
	for _, serverID := range gconf.Get("router.addrs").Keys() {
		addr, err := gconf.Get("router.addrs").String(serverID)
		if err != nil {
			return err
		}
		Conf.RouterRPCAddrs[serverID] = addr
	}
	return nil
}

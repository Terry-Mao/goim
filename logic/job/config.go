package main

import (
	"flag"
	"runtime"
	"strconv"
	"time"

	"github.com/Terry-Mao/goconf"
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
	Log        string   `goconf:"base:log"`
	ZKAddrs    []string `goconf:"kafka:zookeeper.list:,"`
	ZKRoot     string   `goconf:"kafka:zkroot"`
	KafkaTopic string   `goconf:"kafka:topic"`
	// comet
	Comets      map[int32]string `goconf:"-"`
	RoutineSize int64            `goconf:"comet:routine.size"`
	RoutineChan int              `goconf:"comet:routine.chan"`
	// push
	PushChan     int `goconf:"push:chan"`
	PushChanSize int `goconf:"push:chan.size"`
	// timer
	Timer     int `goconf:"timer:num"`
	TimerSize int `goconf:"timer:size"`
	// room
	RoomBatch  int           `goconf:"room:batch"`
	RoomSignal time.Duration `goconf:"room:signal:time"`
	// monitor
	MonitorOpen  bool     `goconf:"monitor:open"`
	MonitorAddrs []string `goconf:"monitor:addrs:,"`
}

func NewConfig() *Config {
	return &Config{
		Comets:       make(map[int32]string),
		ZKRoot:       "",
		KafkaTopic:   "kafka_topic_push",
		RoutineSize:  16,
		RoutineChan:  64,
		PushChan:     4,
		PushChanSize: 100,
		//timer
		// timer
		Timer:     runtime.NumCPU(),
		TimerSize: 1000,
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
	return
}

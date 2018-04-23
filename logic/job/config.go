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
	PprofAddrs []string `goconf:"base:pprof.addrs:,"`
	ZKAddrs    []string `goconf:"kafka:zookeeper.list:,"`
	ZKRoot     string   `goconf:"kafka:zkroot"`
	KafkaGroup string   `goconf:"kafka:group"`
	KafkaTopic string   `goconf:"kafka:topic"`
	// comet
	Comets      map[int32]string `goconf:"-"`
	RoutineSize uint64           `goconf:"comet:routine.size"`
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
	RoomIdle   time.Duration `goconf:"room:idle:time"`
	// monitor
	MonitorOpen  bool     `goconf:"monitor:open"`
	MonitorAddrs []string `goconf:"monitor:addrs:,"`
}

func NewConfig() *Config {
	return &Config{
		Comets:       make(map[int32]string),
		ZKRoot:       "",
		KafkaGroup:   "kafka_topic_push_group",
		KafkaTopic:   "KafkaPushsTopic",
		RoutineSize:  16,
		RoutineChan:  64,
		PushChan:     4,
		PushChanSize: 100,
		// room
		RoomBatch:  40,
		RoomSignal: time.Second,
		RoomIdle:   time.Hour,
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

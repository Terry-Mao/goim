// Copyright © 2014 Terry Mao, LiuDing All rights reserved.
// This file is part of gopush-cluster.

// gopush-cluster is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// gopush-cluster is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with gopush-cluster.  If not, see <http://www.gnu.org/licenses/>.

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
	flag.StringVar(&confFile, "c", "./logic.conf", " set logic config file path")
}

type Config struct {
	// base section
	PidFile          string        `goconf:"base:pidfile"`
	Dir              string        `goconf:"base:dir"`
	Log              string        `goconf:"base:log"`
	MaxProc          int           `goconf:"base:maxproc"`
	PprofAddrs       []string      `goconf:"base:pprof.addrs:,"`
	RPCAddrs         []string      `goconf:"base:rpc.addrs:,"`
	HTTPAddrs        []string      `goconf:"base:http.addrs:,"`
	HTTPReadTimeout  time.Duration `goconf:"base:http.read.timeout:time"`
	HTTPWriteTimeout time.Duration `goconf:"base:http.write.timeout:time"`
	// router RPC
	RouterRPCAddrs map[string]string `-`
	// kafka
	KafkaAddrs []string `goconf:"kafka:addrs"`
	// monitor
	MonitorOpen  bool     `goconf:"monitor:open"`
	MonitorAddrs []string `goconf:"monitor:addrs:,"`
}

func NewConfig() *Config {
	return &Config{
		// base section
		PidFile:        "/tmp/goim-logic.pid",
		Dir:            "./",
		Log:            "./logic-log.xml",
		MaxProc:        runtime.NumCPU(),
		PprofAddrs:     []string{"localhost:6971"},
		HTTPAddrs:      []string{"7172"},
		RouterRPCAddrs: make(map[string]string),
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
	for _, serverID := range gconf.Get("router.addrs").Keys() {
		addr, err := gconf.Get("router.addrs").String(serverID)
		if err != nil {
			return err
		}
		Conf.RouterRPCAddrs[serverID] = addr
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

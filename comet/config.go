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
	flag.StringVar(&confFile, "c", "./comet.conf", " set comet config file path")
}

type Config struct {
	// base section
	PidFile   string   `goconf:"base:pidfile"`
	Dir       string   `goconf:"base:dir"`
	Log       string   `goconf:"base:log"`
	MaxProc   int      `goconf:"base:maxproc"`
	PprofBind []string `goconf:"base:pprof.bind:,"`
	StatBind  []string `goconf:"base:stat.bind:,"`
	ServerId  int32    `goconf:"base:server.id"`
	Debug     bool     `goconf:"base:debug"`
	Whitelist []string `goconf:"base:white.list:,"`
	WhiteLog  string   `goconf:"base:white.log"`
	// tcp
	TCPBind         []string `goconf:"tcp:bind:,"`
	TCPSndbuf       int      `goconf:"tcp:sndbuf:memory"`
	TCPRcvbuf       int      `goconf:"tcp:rcvbuf:memory"`
	TCPKeepalive    bool     `goconf:"tcp:keepalive"`
	TCPReader       int      `goconf:"tcp:reader"`
	TCPReadBuf      int      `goconf:"tcp:readbuf"`
	TCPReadBufSize  int      `goconf:"tcp:readbuf.size"`
	TCPWriter       int      `goconf:"tcp:writer"`
	TCPWriteBuf     int      `goconf:"tcp:writebuf"`
	TCPWriteBufSize int      `goconf:"tcp:writebuf.size"`
	// websocket
	WebsocketBind        []string `goconf:"websocket:bind:,"`
	WebsocketTLSOpen     bool     `goconf:"websocket:tls.open"`
	WebsocketTLSBind     []string `goconf:"websocket:tls.bind:,"`
	WebsocketCertFile    string   `goconf:"websocket:cert.file"`
	WebsocketPrivateFile string   `goconf:"websocket:private.file"`
	// flash safe policy
	FlashPolicyOpen bool     `goconf:"flash:policy.open"`
	FlashPolicyBind []string `goconf:"flash:policy.bind:,"`
	// proto section
	HandshakeTimeout time.Duration `goconf:"proto:handshake.timeout:time"`
	WriteTimeout     time.Duration `goconf:"proto:write.timeout:time"`
	SvrProto         int           `goconf:"proto:svr.proto"`
	CliProto         int           `goconf:"proto:cli.proto"`
	// timer
	Timer     int `goconf:"timer:num"`
	TimerSize int `goconf:"timer:size"`
	// bucket
	Bucket        int   `goconf:"bucket:num"`
	BucketChannel int   `goconf:"bucket:channel"`
	BucketRoom    int   `goconf:"bucket:room"`
	RoutineAmount int64 `goconf:"bucket:routine.amount"`
	RoutineSize   int   `goconf:"bucket:routine.size"`
	// push
	RPCPushAddrs []string `goconf:"push:rpc.addrs:,"`
	// logic
	LogicAddrs []string `goconf:"logic:rpc.addrs:,"`
	// monitor
	MonitorOpen  bool     `goconf:"monitor:open"`
	MonitorAddrs []string `goconf:"monitor:addrs:,"`
}

func NewConfig() *Config {
	return &Config{
		// base section
		PidFile:   "/tmp/goim-comet.pid",
		Dir:       "./",
		Log:       "./comet-log.xml",
		MaxProc:   runtime.NumCPU(),
		PprofBind: []string{"localhost:6971"},
		StatBind:  []string{"localhost:6972"},
		Debug:     true,
		// tcp
		TCPBind:      []string{"0.0.0.0:8080"},
		TCPSndbuf:    1024,
		TCPRcvbuf:    1024,
		TCPKeepalive: false,
		// websocket
		WebsocketBind: []string{"0.0.0.0:8090"},
		// websocket tls
		WebsocketTLSOpen:     false,
		WebsocketTLSBind:     []string{"0.0.0.0:8095"},
		WebsocketCertFile:    "../source/cert.pem",
		WebsocketPrivateFile: "../source/private.pem",
		// flash safe policy
		FlashPolicyOpen: false,
		FlashPolicyBind: []string{"0.0.0.0:843"},
		// proto section
		HandshakeTimeout: 5 * time.Second,
		WriteTimeout:     5 * time.Second,
		TCPReadBuf:       1024,
		TCPWriteBuf:      1024,
		TCPReadBufSize:   1024,
		TCPWriteBufSize:  1024,
		// timer
		Timer:     runtime.NumCPU(),
		TimerSize: 1000,
		// bucket
		Bucket:        1024,
		CliProto:      5,
		SvrProto:      80,
		BucketChannel: 1024,
		// push
		RPCPushAddrs: []string{"localhost:8083"},
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

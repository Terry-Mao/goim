package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/Bilibili/discovery/naming"
	resolver "github.com/Bilibili/discovery/naming/grpc"
	"github.com/Terry-Mao/goim/internal/logic"
	"github.com/Terry-Mao/goim/internal/logic/conf"
	log "github.com/golang/glog"
)

const (
	ver = "2.0.0"
)

func main() {
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	log.Infof("goim-logic [version: %s] start", ver)
	// grpc register naming
	dis := naming.New(conf.Conf.Discovery)
	resolver.Register(dis)
	// logic
	srv := logic.New(conf.Conf)
	// TODO http&grpc
	// signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Infof("goim-logic get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			log.Infof("goim-logic [version: %s] exit", ver)
			srv.Close()
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}

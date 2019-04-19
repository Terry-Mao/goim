package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/bilibili/discovery/naming"

	"github.com/Terry-Mao/goim/internal/job"
	"github.com/Terry-Mao/goim/internal/job/conf"

	resolver "github.com/bilibili/discovery/naming/grpc"
	log "github.com/golang/glog"
)

var (
	ver = "2.0.0"
)

func main() {
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	log.Infof("goim-job [version: %s env: %+v] start", ver, conf.Conf.Env)
	// grpc register naming

	cfgNaming := &naming.Config{
		Nodes:  conf.Conf.Discovery.Nodes,
		Region: conf.Conf.Discovery.Region,
		Zone:   conf.Conf.Discovery.Zone,
		Env:    conf.Conf.Discovery.Env,
		Host:   conf.Conf.Discovery.Host,
	}
	dis := naming.New(cfgNaming)
	resolver.Register(dis)
	// job
	j := job.New(conf.Conf)

	go j.Consume()
	// signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Infof("goim-job get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			_ = j.Close()
			log.Infof("goim-job [version: %s] exit", ver)
			log.Flush()
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}

package main

import (
	"context"
	"flag"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/bilibili/discovery/naming"
	resolver "github.com/bilibili/discovery/naming/grpc"
	"github.com/Terry-Mao/goim/internal/logic"
	"github.com/Terry-Mao/goim/internal/logic/conf"
	"github.com/Terry-Mao/goim/internal/logic/grpc"
	"github.com/Terry-Mao/goim/internal/logic/http"
	"github.com/Terry-Mao/goim/internal/logic/model"
	"github.com/Terry-Mao/goim/pkg/ip"
	log "github.com/golang/glog"
)

const (
	ver   = "2.0.0"
	appid = "goim.logic"
)

func main() {
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	log.Infof("goim-logic [version: %s env: %+v] start", ver, conf.Conf.Env)
	// grpc register naming
	dis := naming.New(conf.Conf.Discovery)
	resolver.Register(dis)
	// logic
	srv := logic.New(conf.Conf)
	httpSrv := http.New(conf.Conf.HTTPServer, srv)
	rpcSrv := grpc.New(conf.Conf.RPCServer, srv)
	cancel := register(dis, srv)
	// signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Infof("goim-logic get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			if cancel != nil {
				cancel()
			}
			srv.Close()
			httpSrv.Close()
			rpcSrv.GracefulStop()
			log.Infof("goim-logic [version: %s] exit", ver)
			log.Flush()
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}

func register(dis *naming.Discovery, srv *logic.Logic) context.CancelFunc {
	env := conf.Conf.Env
	addr := ip.InternalIP()
	_, port, _ := net.SplitHostPort(conf.Conf.RPCServer.Addr)
	ins := &naming.Instance{
		Region:   env.Region,
		Zone:     env.Zone,
		Env:      env.DeployEnv,
		Hostname: env.Host,
		AppID:    appid,
		Addrs: []string{
			"grpc://" + addr + ":" + port,
		},
		Metadata: map[string]string{
			model.MetaWeight: strconv.FormatInt(env.Weight, 10),
		},
	}
	cancel, err := dis.Register(ins)
	if err != nil {
		panic(err)
	}
	return cancel
}

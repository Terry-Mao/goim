package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bilibili/discovery/naming"
	resolver "github.com/bilibili/discovery/naming/grpc"
	"github.com/Terry-Mao/goim/internal/comet"
	"github.com/Terry-Mao/goim/internal/comet/conf"
	"github.com/Terry-Mao/goim/internal/comet/grpc"
	md "github.com/Terry-Mao/goim/internal/logic/model"
	"github.com/Terry-Mao/goim/pkg/ip"
	log "github.com/golang/glog"
)

const (
	ver   = "2.0.0"
	appid = "goim.comet"
)

func main() {
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	rand.Seed(time.Now().UTC().UnixNano())
	runtime.GOMAXPROCS(runtime.NumCPU())
	println(conf.Conf.Debug)
	log.Infof("goim-comet [version: %s env: %+v] start", ver, conf.Conf.Env)
	// register discovery
	dis := naming.New(conf.Conf.Discovery)
	resolver.Register(dis)
	// new comet server
	srv := comet.NewServer(conf.Conf)
	if err := comet.InitWhitelist(conf.Conf.Whitelist); err != nil {
		panic(err)
	}
	if err := comet.InitTCP(srv, conf.Conf.TCP.Bind, runtime.NumCPU()); err != nil {
		panic(err)
	}
	if err := comet.InitWebsocket(srv, conf.Conf.Websocket.Bind, runtime.NumCPU()); err != nil {
		panic(err)
	}
	if conf.Conf.Websocket.TLSOpen {
		if err := comet.InitWebsocketWithTLS(srv, conf.Conf.Websocket.TLSBind, conf.Conf.Websocket.CertFile, conf.Conf.Websocket.PrivateFile, runtime.NumCPU()); err != nil {
			panic(err)
		}
	}
	// new grpc server
	rpcSrv := grpc.New(conf.Conf.RPCServer, srv)
	cancel := register(dis, srv)
	// signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Infof("goim-comet get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			if cancel != nil {
				cancel()
			}
			rpcSrv.GracefulStop()
			srv.Close()
			log.Infof("goim-comet [version: %s] exit", ver)
			log.Flush()
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}

func register(dis *naming.Discovery, srv *comet.Server) context.CancelFunc {
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
			md.MetaWeight:  strconv.FormatInt(env.Weight, 10),
			md.MetaOffline: strconv.FormatBool(env.Offline),
			md.MetaAddrs:   strings.Join(env.Addrs, ","),
		},
	}
	cancel, err := dis.Register(ins)
	if err != nil {
		panic(err)
	}
	// renew discovery metadata
	go func() {
		for {
			var (
				err   error
				conns int
				ips   = make(map[string]struct{})
			)
			for _, bucket := range srv.Buckets() {
				for ip := range bucket.IPCount() {
					ips[ip] = struct{}{}
				}
				conns += bucket.ChannelCount()
			}
			ins.Metadata[md.MetaConnCount] = fmt.Sprint(conns)
			ins.Metadata[md.MetaIPCount] = fmt.Sprint(len(ips))
			if err = dis.Set(ins); err != nil {
				log.Errorf("dis.Set(%+v) error(%v)", ins, err)
				time.Sleep(time.Second)
				continue
			}
			time.Sleep(time.Second * 10)
		}
	}()
	return cancel
}

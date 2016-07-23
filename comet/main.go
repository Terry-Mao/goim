package main

import (
	"flag"
	"goim/libs/perf"
	"runtime"

	log "github.com/thinkboy/log4go"
)

var (
	DefaultServer *Server
	Debug         bool
)

func main() {
	flag.Parse()
	if err := InitConfig(); err != nil {
		panic(err)
	}
	Debug = Conf.Debug
	runtime.GOMAXPROCS(Conf.MaxProc)
	log.LoadConfiguration(Conf.Log)
	defer log.Close()
	log.Info("comet[%s] start", Ver)
	// white list log
	if err := InitWhiteList(Conf.WhiteLog); err != nil {
		panic(err)
	}
	perf.Init(Conf.PprofBind)
	// logic rpc
	if err := InitLogicRpc(Conf.LogicAddr); err != nil {
		log.Warn("logic rpc current can't connect, retry")
	}
	// new server
	buckets := make([]*Bucket, Conf.Bucket)
	for i := 0; i < Conf.Bucket; i++ {
		buckets[i] = NewBucket(BucketOptions{
			ChannelSize:   Conf.BucketChannel,
			RoomSize:      Conf.BucketRoom,
			RoutineAmount: Conf.RoutineAmount,
			RoutineSize:   Conf.RoutineSize,
		})
	}
	round := NewRound(RoundOptions{
		Reader:       Conf.TCPReader,
		ReadBuf:      Conf.TCPReadBuf,
		ReadBufSize:  Conf.TCPReadBufSize,
		Writer:       Conf.TCPWriter,
		WriteBuf:     Conf.TCPWriteBuf,
		WriteBufSize: Conf.TCPWriteBufSize,
		Timer:        Conf.Timer,
		TimerSize:    Conf.TimerSize,
	})
	operator := new(DefaultOperator)
	DefaultServer = NewServer(buckets, round, operator, ServerOptions{
		CliProto:         Conf.CliProto,
		SvrProto:         Conf.SvrProto,
		HandshakeTimeout: Conf.HandshakeTimeout,
		TCPKeepalive:     Conf.TCPKeepalive,
		TCPRcvbuf:        Conf.TCPRcvbuf,
		TCPSndbuf:        Conf.TCPSndbuf,
	})
	// white list
	DefaultServer.Whitelist = make(map[string]struct{})
	for _, key := range Conf.Whitelist {
		DefaultServer.Whitelist[key] = struct{}{}
	}
	// tcp comet
	if err := InitTCP(Conf.TCPBind, Conf.MaxProc); err != nil {
		panic(err)
	}
	// websocket comet
	if err := InitWebsocket(Conf.WebsocketBind); err != nil {
		panic(err)
	}
	// flash safe policy
	if Conf.FlashPolicyOpen {
		if err := InitFlashPolicy(); err != nil {
			panic(err)
		}
	}
	// wss comet
	if Conf.WebsocketTLSOpen {
		if err := InitWebsocketWithTLS(Conf.WebsocketTLSBind, Conf.WebsocketCertFile, Conf.WebsocketPrivateFile); err != nil {
			panic(err)
		}
	}
	// start rpc
	if err := InitRPCPush(Conf.RPCPushAddrs); err != nil {
		panic(err)
	}
	// block until a signal is received.
	InitSignal()
}

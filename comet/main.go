package main

import (
	log "code.google.com/p/log4go"
	"flag"
	"github.com/Terry-Mao/goim/libs/perf"
	"runtime"
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
	perf.Init(Conf.PprofBind)
	// logic rpc
	if err := InitLogicRpc(Conf.LogicAddr); err != nil {
		log.Warn("logic rpc current can't connect, retry")
	}
	// new server
	buckets := make([]*Bucket, Conf.Bucket)
	for i := 0; i < Conf.Bucket; i++ {
		buckets[i] = NewBucket(BucketOptions{
			ChannelSize: Conf.BucketChannel,
			RoomSize:    Conf.BucketRoom,
		}, RoomOptions{
			ChannelSize: Conf.RoomChannel,
			ProtoSize:   Conf.RoomProto,
			BatchNum:    Conf.RoomBatch,
			SignalTime:  Conf.RoomSignal,
		})
	}
	round := NewRound(Conf.TCPReadBuf, Conf.TCPWriteBuf, Conf.Timer, Conf.TimerSize)
	operator := new(DefaultOperator)
	DefaultServer = NewServer(buckets, round, operator, ServerOptions{
		SvrProto:         Conf.SvrProto,
		HandshakeTimeout: Conf.HandshakeTimeout,
		TCPKeepalive:     Conf.TCPKeepalive,
		TCPRcvbuf:        Conf.TCPRcvbuf,
		TCPSndbuf:        Conf.TCPSndbuf,
		TCPReadBufSize:   Conf.TCPReadBufSize,
		TCPWriteBufSize:  Conf.TCPWriteBufSize,
	})
	// tcp comet
	if err := InitTCP(Conf.TCPBind, Conf.MaxProc); err != nil {
		panic(err)
	}
	// websocket comet
	if err := InitWebsocket(Conf.WebsocketBind); err != nil {
		panic(err)
	}
	// wss comet
	if Conf.WebsocketTLSOpen {
		if err := InitWebsocketWithTLS(Conf.WebsocketTLSBind, Conf.WebsocketCertFile, Conf.WebsocketPrivateFile); err != nil {
			panic(err)
		}
	}
	// http comet
	if err := InitHTTP(Conf.HTTPBind); err != nil {
		panic(err)
	}
	// start rpc
	if err := InitRPCPush(Conf.RPCPushAddrs); err != nil {
		panic(err)
	}
	// block until a signal is received.
	InitSignal()
}

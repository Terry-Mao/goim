package main

import (
	log "code.google.com/p/log4go"
	"flag"
	"github.com/Terry-Mao/goim/libs/perf"
	"runtime"
)

var (
	DefaultServer *Server
)

func main() {
	flag.Parse()
	if err := InitConfig(); err != nil {
		panic(err)
	}
	runtime.GOMAXPROCS(Conf.MaxProc)
	log.LoadConfiguration(Conf.Log)
	defer log.Close()
	log.Info("comet[%s] start", Ver)
	perf.Init(Conf.PprofBind)
	if err := InitRSA(); err != nil {
		panic(err)
	}
	// new server
	buckets := make([]*Bucket, Conf.Bucket)
	for i := 0; i < Conf.Bucket; i++ {
		buckets[i] = NewBucket(Conf.Channel, Conf.CliProto, Conf.SvrProto)
	}
	round := NewRound(Conf.ReadBuf, Conf.WriteBuf, Conf.Timer, Conf.TimerSize, Conf.Session, Conf.SessionSize)
	codec := new(BinaryServerCodec)
	operator := new(DefaultOperator)
	cryptor := NewDefaultCryptor()
	DefaultServer = NewServer(buckets, round, codec, operator, cryptor)
	if err := InitTCP(); err != nil {
		panic(err)
	}
	if err := InitWebsocket(); err != nil {
		panic(err)
	}
	if err := InitHttpPush(); err != nil {
		panic(err)
	}
	// start rpc
	if err := InitRPCPush(); err != nil {
		panic(err)
	}
	// block until a signal is received.
	InitSignal()
}

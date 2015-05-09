package main

import (
	log "code.google.com/p/log4go"
	"sync"
)

type Round struct {
	readers   []*sync.Pool
	writers   []*sync.Pool
	timers    []*Timer
	protos    []*FreeProto
	readerIdx int
	writerIdx int
	timerIdx  int
	protoIdx  int
}

func NewRound(readBuf, writeBuf, timer, timerSize, proto, protoSize int) *Round {
	r := new(Round)
	log.Debug("create %d reader buffer pool", readBuf)
	r.readerIdx = readBuf - 1
	r.readers = make([]*sync.Pool, readBuf)
	for i := 0; i < readBuf; i++ {
		r.readers[i] = new(sync.Pool)
	}
	log.Debug("create %d writer buffer pool", writeBuf)
	r.writerIdx = writeBuf - 1
	r.writers = make([]*sync.Pool, writeBuf)
	for i := 0; i < writeBuf; i++ {
		r.writers[i] = new(sync.Pool)
	}
	log.Debug("create %d timer", timer)
	r.timerIdx = timer - 1
	r.timers = make([]*Timer, timer)
	for i := 0; i < timer; i++ {
		r.timers[i] = NewTimer(timerSize)
	}
	log.Debug("create %d free proto", proto)
	r.protoIdx = proto - 1
	r.protos = make([]*FreeProto, proto)
	for i := 0; i < proto; i++ {
		r.protos[i] = NewFreeProto(protoSize)
	}
	return r
}

func (r *Round) Timer(rn int) *Timer {
	return r.timers[rn&r.timerIdx]
}

func (r *Round) Reader(rn int) *sync.Pool {
	return r.readers[rn&r.readerIdx]
}

func (r *Round) Writer(rn int) *sync.Pool {
	return r.writers[rn&r.writerIdx]
}

func (r *Round) Proto(rn int) *FreeProto {
	return r.protos[rn&r.protoIdx]
}

package main

import (
	log "code.google.com/p/log4go"
	"sync"
)

type Round struct {
	readers   []sync.Pool
	writers   []sync.Pool
	timers    []Timer
	readerIdx int
	writerIdx int
	timerIdx  int
}

func NewRound(readBuf, writeBuf, timer, timerSize int) *Round {
	r := new(Round)
	log.Debug("create %d reader buffer pool", readBuf)
	log.Debug("create %d writer buffer pool", writeBuf)
	r.readerIdx = readBuf
	r.writerIdx = writeBuf
	r.readers = make([]sync.Pool, readBuf)
	r.writers = make([]sync.Pool, writeBuf)
	log.Debug("create %d timer", timer)
	r.timerIdx = timer
	r.timers = make([]Timer, timer)
	for i := 0; i < timer; i++ {
		r.timers[i].Init(timerSize)
	}
	return r
}

func (r *Round) Timer(rn int) *Timer {
	return &(r.timers[rn%r.timerIdx])
}

func (r *Round) Reader(rn int) *sync.Pool {
	return &(r.readers[rn%r.readerIdx])
}

func (r *Round) Writer(rn int) *sync.Pool {
	return &(r.writers[rn%r.writerIdx])
}

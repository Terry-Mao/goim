package main

import (
	log "code.google.com/p/log4go"
	"sync"
)

type Round struct {
	readers []*sync.Pool
	writers []*sync.Pool
	//rpackers   []*sync.Pool
	//wpackers   []*sync.Pool
	timers    []*Timer
	readerIdx int
	writerIdx int
	//rpackerIdx int
	//wpackerIdx int
	timerIdx int
}

func NewRound(readBuf, writeBuf, timer, timerSize int) *Round {
	r := new(Round)
	log.Debug("create %d reader buffer pool", readBuf)
	r.readerIdx = readBuf
	r.readers = make([]*sync.Pool, readBuf)
	for i := 0; i < readBuf; i++ {
		r.readers[i] = new(sync.Pool)
	}
	log.Debug("create %d writer buffer pool", writeBuf)
	r.writerIdx = writeBuf
	r.writers = make([]*sync.Pool, writeBuf)
	for i := 0; i < writeBuf; i++ {
		r.writers[i] = new(sync.Pool)
	}
	/*
		log.Debug("create %d read pack buffer pool", readBuf)
		r.rpackerIdx = readBuf
		r.rpackers = make([]*sync.Pool, readBuf)
		for i := 0; i < readBuf; i++ {
			r.rpackers[i] = new(sync.Pool)
		}
		log.Debug("create %d writer pack buffer pool", writeBuf)
		r.wpackerIdx = writeBuf
		r.wpackers = make([]*sync.Pool, writeBuf)
		for i := 0; i < writeBuf; i++ {
			r.wpackers[i] = new(sync.Pool)
		}
	*/
	log.Debug("create %d timer", timer)
	r.timerIdx = timer
	r.timers = make([]*Timer, timer)
	for i := 0; i < timer; i++ {
		r.timers[i] = NewTimer(timerSize)
	}
	// start timer process
	go TimerProcess(r.timers)
	return r
}

func (r *Round) Timer(rn int) *Timer {
	return r.timers[rn%r.timerIdx]
}

func (r *Round) Reader(rn int) *sync.Pool {
	return r.readers[rn%r.readerIdx]
}

func (r *Round) Writer(rn int) *sync.Pool {
	return r.writers[rn%r.writerIdx]
}

/*
func (r *Round) Rpacker(rn int) *sync.Pool {
	return r.rpackers[rn%r.rpackerIdx]
}

func (r *Round) Wpacker(rn int) *sync.Pool {
	return r.wpackers[rn%r.wpackerIdx]
}
*/

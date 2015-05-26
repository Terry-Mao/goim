package main

import (
	log "code.google.com/p/log4go"
	"sync"
)

type Round struct {
	readers []*sync.Pool
	writers []*sync.Pool
	//encrypters   []*sync.Pool
	//decrypters   []*sync.Pool
	timers   []*Timer
	sessions []*Session
	// protos    []*FreeProto
	readerIdx int
	writerIdx int
	//encrypterIdx int
	//decrypterIdx int
	timerIdx   int
	sessionIdx int
	// protoIdx int
}

func NewRound(readBuf, writeBuf, timer, timerSize, session, sessionSize int) *Round {
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
	log.Debug("create %d timer", timer)
	r.timerIdx = timer
	r.timers = make([]*Timer, timer)
	for i := 0; i < timer; i++ {
		r.timers[i] = NewTimer(timerSize)
	}
	// start timer process
	go TimerProcess(r.timers)
	log.Debug("create %d session", session)
	r.sessionIdx = session
	r.sessions = make([]*Session, session)
	for i := 0; i < session; i++ {
		r.sessions[i] = NewSession(sessionSize)
	}
	go SessionProcess(r.sessions)
	/*
		log.Debug("create %d encrypter buffer pool", encrypterBuf)
		r.encrypterIdx = encrypterBuf - 1
		r.encrypters = make([]*sync.Pool, encrypterBuf)
		for i := 0; i < encrypterBuf; i++ {
			r.encrypters[i] = new(sync.Pool)
		}
		log.Debug("create %d decrypter buffer pool", decrypterBuf)
		r.decrypterIdx = decrypterBuf - 1
		r.decrypters = make([]*sync.Pool, decrypterBuf)
		for i := 0; i < encrypterBuf; i++ {
			r.decrypters[i] = new(sync.Pool)
		}
		log.Debug("create %d free proto", proto)
		r.protoIdx = proto
		r.protos = make([]*FreeProto, proto)
		for i := 0; i < proto; i++ {
			r.protos[i] = NewFreeProto(protoSize)
		}
	*/
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

func (r *Round) Session(rn int) *Session {
	return r.sessions[rn%r.sessionIdx]
}

/*
func (r *Round) Proto(rn int) *FreeProto {
	return r.protos[rn%r.protoIdx]
}

func (r *Round) Encrypter(rn int) *sync.Pool {
	return r.encrypters[rn&r.encrypterIdx]
}

func (r *Round) Decrypter(rn int) *sync.Pool {
	return r.decrypters[rn&r.decrypterIdx]
}
*/

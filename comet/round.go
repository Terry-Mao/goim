package main

import (
	"goim/libs/bytes"
	"goim/libs/time"
)

type RoundOptions struct {
	Timer        int
	TimerSize    int
	Reader       int
	ReadBuf      int
	ReadBufSize  int
	Writer       int
	WriteBuf     int
	WriteBufSize int
}

// Ronnd userd for connection round-robin get a reader/writer/timer for split big lock.
type Round struct {
	readers   []bytes.Pool
	writers   []bytes.Pool
	timers    []time.Timer
	options   RoundOptions
	readerIdx int
	writerIdx int
	timerIdx  int
}

// NewRound new a round struct.
func NewRound(options RoundOptions) (r *Round) {
	var i int
	r = new(Round)
	r.options = options
	// reader
	r.readers = make([]bytes.Pool, options.Reader)
	for i = 0; i < options.Reader; i++ {
		r.readers[i].Init(options.ReadBuf, options.ReadBufSize)
	}
	// writer
	r.writers = make([]bytes.Pool, options.Writer)
	for i = 0; i < options.Writer; i++ {
		r.writers[i].Init(options.WriteBuf, options.WriteBufSize)
	}
	// timer
	r.timers = make([]time.Timer, options.Timer)
	for i = 0; i < options.Timer; i++ {
		r.timers[i].Init(options.TimerSize)
	}
	return
}

// Timer get a timer.
func (r *Round) Timer(rn int) *time.Timer {
	return &(r.timers[rn%r.options.Timer])
}

// Reader get a reader memory buffer.
func (r *Round) Reader(rn int) *bytes.Pool {
	return &(r.readers[rn%r.options.Reader])
}

// Writer get a writer memory buffer pool.
func (r *Round) Writer(rn int) *bytes.Pool {
	return &(r.writers[rn%r.options.Writer])
}

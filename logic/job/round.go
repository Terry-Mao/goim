package main

import (
	"goim/libs/time"
)

type RoundOptions struct {
	Timer     int
	TimerSize int
}

// Ronnd userd for connection round-robin get a timer for split big lock.
type Round struct {
	timers   []time.Timer
	options  RoundOptions
	timerIdx int
}

// NewRound new a round struct.
func NewRound(options RoundOptions) (r *Round) {
	var i int
	r = new(Round)
	r.options = options
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

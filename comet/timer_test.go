package main

import (
	log "code.google.com/p/log4go"
	"testing"
	"time"
)

func TestTimer(t *testing.T) {
	var err error
	timer := NewTimer(100)
	tds := make([]*TimerData, 100)
	for i := 0; i < 100; i++ {
		if tds[i], err = timer.Add(time.Duration(i)*time.Second+5*time.Minute, nil); err != nil {
			t.Error(err)
			t.FailNow()
		}
	}
	// overflow
	if _, err = timer.Add(1*time.Second, nil); err == nil {
		t.Error(err)
		t.FailNow()
	}
	printTimer(timer)
	for i := 0; i < 100; i++ {
		log.Debug("td: %s, %s, %d", tds[i].Key, tds[i].ExpireString(), tds[i].index)
		timer.Del(tds[i])
	}
	printTimer(timer)
	for i := 0; i < 100; i++ {
		if tds[i], err = timer.Add(time.Duration(i)*time.Second+5*time.Minute, nil); err != nil {
			t.Error(err)
			t.FailNow()
		}
	}
	printTimer(timer)
	for i := 0; i < 100; i++ {
		timer.del(0)
	}
	printTimer(timer)
}

func TestTimerProcess(t *testing.T) {
	// process test
	timers := make([]Timer, 1)
	timers[0].Init(3)
	timer := &(timers[0])
	timerd, err := timer.Add(5*time.Second, nil)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	go TimerProcess(timers)
	time.Sleep(10 * time.Second)
	timer.Del(timerd)
}

func printTimer(timer *Timer) {
	log.Debug("--------------------")
	for i := 0; i <= timer.cur; i++ {
		log.Debug("timer: %s, %s, index: %d", timer.timers[i].Key, timer.timers[i].ExpireString(), timer.timers[i].index)
	}
	log.Debug("--------------------")
}

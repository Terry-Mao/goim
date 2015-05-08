package main

import (
	log "code.google.com/p/log4go"
	"testing"
	"time"
)

func TestTimer(t *testing.T) {
	timer := NewTimer(100)
	now := time.Now().Add(5 * time.Minute)
	tds := make([]*TimerData, 100)
	for i := 0; i < 100; i++ {
		tds[i] = new(TimerData)
		tds[i].key = time.Now().Add(5 * time.Minute).Add(time.Duration(i) * time.Second)
		if err := timer.Add(tds[i]); err != nil {
			t.Error(err)
			t.FailNow()
		}
	}
	td := new(TimerData)
	td.key = now
	// overflow
	if err := timer.Add(td); err == nil {
		t.Error(err)
		t.FailNow()
	}
	printTimer(timer)
	for i := 0; i < 100; i++ {
		log.Debug("td: %s, %d", tds[i].String(), tds[i].index)
		if err := timer.Del(tds[i]); err != nil {
			t.Error(err)
			t.FailNow()
		}
	}
	printTimer(timer)
	for i := 0; i < 100; i++ {
		tds[i] = new(TimerData)
		tds[i].key = now
		if err := timer.Add(tds[i]); err != nil {
			t.Error(err)
			t.FailNow()
		}
	}
	printTimer(timer)
	for i := 0; i < 100; i++ {
		if _, err := timer.remove(0); err != nil {
			t.Error(err)
			t.FailNow()
		}
	}
	printTimer(timer)
}

func TestTimerProcess(t *testing.T) {
	// process test
	td := new(TimerData)
	timer := NewTimer(3)
	td.Set(5*time.Second, nil)
	if err := timer.Add(td); err != nil {
		t.Error(err)
		t.FailNow()
	}
	time.Sleep(10 * time.Second)
}

func printTimer(timer *Timer) {
	log.Debug("--------------------")
	for i := 0; i <= timer.cur; i++ {
		log.Debug("timer: %s, index: %d", timer.timers[i].String(), timer.timers[i].index)
	}
	log.Debug("--------------------")
}

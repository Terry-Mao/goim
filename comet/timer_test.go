package main

import (
	log "code.google.com/p/log4go"
	"testing"
	"time"
)

func TestTimer(t *testing.T) {
	timer := NewTimer(3)
	timerData1 := new(TimerData)
	timerData1.key = time.Now()
	if err := timer.Push(timerData1); err != nil {
		t.Error(err)
		t.FailNow()
	}
	timerData2 := new(TimerData)
	timerData2.key = time.Now()
	if err := timer.Push(timerData2); err != nil {
		t.Error(err)
		t.FailNow()
	}
	timerData3 := new(TimerData)
	timerData3.key = time.Now()
	if err := timer.Push(timerData3); err != nil {
		t.Error(err)
		t.FailNow()
	}
	// remove 2
	if _, err := timer.Remove(timerData2); err != nil {
		t.Error(err)
		t.FailNow()
	}
	// re add 2
	if err := timer.Push(timerData2); err != nil {
		t.Error(err)
		t.FailNow()
	}
	printTimer(timer)
	if _, err := timer.Pop(); err != nil {
		t.Error(err)
		t.FailNow()
	}
	if _, err := timer.Pop(); err != nil {
		t.Error(err)
		t.FailNow()
	}
	if _, err := timer.Pop(); err != nil {
		t.Error(err)
		t.FailNow()
	}
	if _, err := timer.Pop(); err == nil {
		t.Error(err)
		t.FailNow()
	}
	printTimer(timer)
}

func TestTimerProcess(t *testing.T) {
	// process test
	var timerData TimerData
	timer := NewTimer(3)
	timerData.Set(5*time.Second, nil)
	if err := timer.Push(&timerData); err != nil {
		t.Error(err)
		t.FailNow()
	}
	time.Sleep(10 * time.Second)
}

func printTimer(timer *Timer) {
	for i := 0; i <= timer.cur; i++ {
		log.Debug("timer : %s", timer.timers[i].key.Format("2006-01-02 15:04:05"))
	}
}

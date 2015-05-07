package main

import (
	log "code.google.com/p/log4go"
	"testing"
	"time"
)

func TestTimer(t *testing.T) {
	timer := NewTimer(3)
	timerData1 := new(TimerData)
	timerData1.Key = 1
	if err := timer.Push(timerData1); err != nil {
		t.Error(err)
		t.FailNow()
	}
	timerData2 := new(TimerData)
	timerData2.Key = 2
	if err := timer.Push(timerData2); err != nil {
		t.Error(err)
		t.FailNow()
	}
	timerData3 := new(TimerData)
	timerData3.Key = 3
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
	if timerData, err := timer.Pop(); err != nil {
		t.Error(err)
		t.FailNow()
	} else {
		log.Debug("pop timer: Key: %d", timerData.Key)
	}
	if timerData, err := timer.Pop(); err != nil {
		t.Error(err)
		t.FailNow()
	} else {
		log.Debug("pop timer: Key: %d", timerData.Key)
	}
	if timerData, err := timer.Pop(); err != nil {
		t.Error(err)
		t.FailNow()
	} else {
		log.Debug("pop timer: Key: %d", timerData.Key)
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
	timerData.Key = time.Now().UnixNano() + int64(5*time.Second)
	timerData.Value = nil
	if err := timer.Push(&timerData); err != nil {
		t.Error(err)
		t.FailNow()
	}
	time.Sleep(10 * time.Second)
}

func printTimer(timer *Timer) {
	for i := 0; i <= timer.cur; i++ {
		log.Debug("timer items: %d", timer.timers[i].Key)
	}
}

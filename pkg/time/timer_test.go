package time

import (
	"fmt"
	"testing"
	"time"
)

func TestTimer(t *testing.T) {
	n := 5
	timer := NewTimer(n)
	tds := make([]*TimerData, n)
	for i := 0; i < n; i++ {
		tds[i] = timer.Add(time.Duration(i)*time.Second+5*time.Minute, nil)
	}
	printTimer(timer)
	for i := 0; i < n; i++ {
		fmt.Printf("td[%d]: %s, %s, %d\n", i, tds[i].Key, tds[i].ExpireString(), tds[i].index)
		timer.Del(tds[i])
	}
	printTimer(timer)
	for i := 0; i < n; i++ {
		tds[i] = timer.Add(time.Duration(i)*time.Second+5*time.Minute, nil)
	}
	printTimer(timer)
	for i := 0; i < n; i++ {
		timer.Del(tds[i])
	}
	printTimer(timer)
	timer.Add(time.Second, nil)
	time.Sleep(time.Second * 2)
	if len(timer.timers) != 0 {
		t.FailNow()
	}
}

func printTimer(timer *Timer) {
	fmt.Printf("----------timers: %d ----------\n", len(timer.timers))
	for i := 0; i < len(timer.timers); i++ {
		fmt.Printf("timers[%d]: %s, %s, index: %d\n", i, timer.timers[i].Key, timer.timers[i].ExpireString(), timer.timers[i].index)
	}
	fmt.Printf("--------------------\n")
}

func Test_TimeReset(t *testing.T) {
	t1 := time.NewTimer(infiniteDuration)
	fmt.Printf("%v start\n", curTime())

	go func() {
		t1.Reset(2 * time.Second)
	}()
	//time.Sleep(1 * time.Second)
	<-t1.C
	fmt.Printf("%v end\n", curTime())
}

func curTime() string {
	return time.Now().Format(timerFormat)
}

package main

import (
	log "code.google.com/p/log4go"
	"net"
	"sync"
	"time"
)

const (
	timerDelay = 100 * time.Millisecond
)

type TimerData struct {
	key   time.Time
	value net.Conn
	index int
}

func (td *TimerData) Set(expire time.Duration, value net.Conn) {
	td.key = time.Now().Add(expire)
	td.value = value
}

func (td *TimerData) String() string {
	return td.key.Format("2006-01-02 15:04:05")
}

type Timer struct {
	cur    int
	max    int
	lock   sync.Mutex
	timers []*TimerData
}

// A heap must be initialized before any of the heap operations
// can be used. Init is idempotent with respect to the heap invariants
// and may be called whenever the heap invariants may have been invalidated.
// Its complexity is O(n) where n = h.Len().
//
func NewTimer(num int) *Timer {
	// heapify
	t := new(Timer)
	t.timers = make([]*TimerData, num, num)
	t.cur = -1
	t.max = num - 1
	go t.process()
	return t
}

// Push pushes the element x onto the heap. The complexity is
// O(log(n)) where n = h.Len().
//
func (t *Timer) Push(item *TimerData) error {
	log.Debug("timer: push item key: %s", item.String())
	log.Debug("timer: before push cur: %d, max: %d", t.cur, t.max)
	t.lock.Lock()
	if t.cur >= t.max {
		t.lock.Unlock()
		return ErrTimerFull
	}
	t.cur++
	item.index = t.cur
	// add to the minheap last node
	t.timers[t.cur] = item
	t.up(t.cur - 1)
	t.lock.Unlock()
	log.Debug("timer: after push cur: %d, max: %d", t.cur, t.max)
	return nil
}

// Pop removes the minimum element (according to Less) from the heap
// and returns it. The complexity is O(log(n)) where n = h.Len().
// It is equivalent to Remove(h, 0).
//
func (t *Timer) Pop() (item *TimerData, err error) {
	log.Debug("timer: before pop cur: %d, max: %d", t.cur, t.max)
	t.lock.Lock()
	if t.cur < 0 {
		t.lock.Unlock()
		err = ErrTimerEmpty
		return
	}
	t.swap(0, t.cur)
	t.down(0, t.cur)
	// remove last element
	item = t.pop()
	t.lock.Unlock()
	log.Debug("timer: after pop cur: %d, max: %d", t.cur, t.max)
	return
}

// Remove removes the element at index i from the heap.
// The complexity is O(log(n)) where n = h.Len().
//
func (t *Timer) Remove(item *TimerData) (nitem *TimerData, err error) {
	log.Debug("timer: remove item key: %s", item.String())
	log.Debug("timer: before remove cur: %d, max: %d", t.cur, t.max)
	t.lock.Lock()
	if item.index == -1 {
		t.lock.Unlock()
		err = ErrTimerNoItem
		return
	}
	if t.cur != item.index {
		// swap the last node
		// let the big one down
		// let the small one up
		t.swap(item.index, t.cur)
		t.down(item.index, t.cur)
		t.up(item.index)
	}
	// remove item is the last node
	nitem = t.pop()
	t.lock.Unlock()
	log.Debug("timer: after remove cur: %d, max: %d", t.cur, t.max)
	return
}

func (t *Timer) Peek() (item *TimerData, err error) {
	if t.cur < 0 {
		err = ErrTimerEmpty
		return
	}
	item = t.timers[0]
	return
}

func (t *Timer) Update(item *TimerData, expire time.Duration) (err error) {
	// item may change in removing
	if item, err = t.Remove(item); err != nil {
		return
	}
	item.key = time.Now().Add(expire)
	err = t.Push(item)
	return
}

func (t *Timer) up(j int) {
	for {
		i := (j - 1) / 2 // parent
		if i == j || !t.less(j, i) {
			break
		}
		t.swap(i, j)
		j = i
	}
}

func (t *Timer) down(i, n int) {
	for {
		j1 := 2*i + 1
		if j1 >= n || j1 < 0 { // j1 < 0 after int overflow
			break
		}
		j := j1 // left child
		if j2 := j1 + 1; j2 < n && !t.less(j1, j2) {
			j = j2 // = 2*i + 2  // right child
		}
		if !t.less(j, i) {
			break
		}
		t.swap(i, j)
		i = j
	}
}

func (t *Timer) less(i, j int) bool {
	return t.timers[i].key.Before(t.timers[j].key)
}

func (t *Timer) swap(i, j int) {
	t.timers[i], t.timers[j] = t.timers[j], t.timers[i]
	t.timers[i].index = i
	t.timers[j].index = j
}

func (t *Timer) pop() (item *TimerData) {
	item = t.timers[t.cur]
	item.index = -1 // for safety
	t.cur--
	return
}

func (t *Timer) process() {
	var (
		err   error
		td    *TimerData
		now   time.Time
		sleep time.Duration
	)
	log.Info("start process timer")
	for {
		if td, err = t.Peek(); err != nil {
			//log.Debug("timer: no expire")
			time.Sleep(timerDelay)
			continue
		}
		now = time.Now()
		if sleep = td.key.Sub(now); int64(sleep) > 0 {
			log.Debug("timer: delay %s", sleep.String())
			time.Sleep(sleep)
			continue
		}
		if td, err = t.Pop(); err != nil {
			time.Sleep(timerDelay)
			continue
		}
		// TODO recheck?
		log.Debug("expire timer: %s", td.String())
		if td.value == nil {
			log.Warn("expire timer no net.Conn")
			continue
		}
		if err = td.value.Close(); err != nil {
			log.Error("timer conn close error(%v)", err)
		}
	}
}

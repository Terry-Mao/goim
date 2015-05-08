package main

import (
	log "code.google.com/p/log4go"
	"net"
	"sync"
	"time"
)

const (
	zeroDuration     = time.Duration(0)
	infiniteDuration = time.Duration(-1)
	timerDelay       = 100 * time.Millisecond
	maxTimerDelay    = 500 * time.Millisecond
)

type TimerData struct {
	key   time.Time
	value net.Conn
	index int
}

func (td *TimerData) Delay() time.Duration {
	return td.key.Sub(time.Now())
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
func (t *Timer) Add(item *TimerData) error {
	t.lock.Lock()
	if t.cur >= t.max {
		t.lock.Unlock()
		return ErrTimerFull
	}
	t.cur++
	item.index = t.cur
	// add to the minheap last node
	t.timers[t.cur] = item
	t.up(t.cur)
	t.lock.Unlock()
	log.Debug("timer: push item key: %s, index: %d", item.String(), item.index)
	return nil
}

// Pop removes the minimum element (according to Less) from the heap
// and returns it. The complexity is O(log(n)) where n = h.Len().
// It is equivalent to Del(0).
//
func (t *Timer) Expire() {
	var (
		err error
		d   time.Duration
	)
	t.lock.Lock()
	for t.cur >= 0 {
		if d = t.timers[0].Delay(); d > 0 {
			break
		}
		if t.timers[0].value == nil {
			log.Warn("expire timer no net.Conn")
		} else {
			if err = t.timers[0].value.Close(); err != nil {
				log.Error("timer conn close error(%v)", err)
			}
		}
		// remove
		if _, err = t.remove(0); err != nil {
			break
		}
	}
	t.lock.Unlock()
	return
}

// Del removes the element at index i from the heap.
// The complexity is O(log(n)) where n = h.Len().
//
func (t *Timer) Del(item *TimerData) (err error) {
	var nitem *TimerData
	t.lock.Lock()
	nitem, err = t.remove(item.index)
	t.lock.Unlock()
	log.Debug("timer: remove item key: %s, index: %d", nitem.String(), nitem.index)
	return
}

func (t *Timer) remove(i int) (nitem *TimerData, err error) {
	if i == -1 {
		log.Error("timer: remove item key: %s not exist", nitem.String())
		err = ErrTimerNoItem
		return
	}
	if t.cur != i {
		t.swap(i, t.cur)
		t.down(i, t.cur)
		t.up(i)
	}
	// remove item is the last node
	nitem = t.last()
	return
}

func (t *Timer) Find() (d time.Duration) {
	t.lock.Lock()
	if t.cur < 0 {
		d = infiniteDuration
		t.lock.Unlock()
		return
	}
	d = t.timers[0].Delay()
	t.lock.Unlock()
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
	//log.Debug("swap(%d, %d)", i, j)
	t.timers[i], t.timers[j] = t.timers[j], t.timers[i]
	t.timers[i].index = i
	t.timers[j].index = j
}

func (t *Timer) last() (item *TimerData) {
	item = t.timers[t.cur]
	item.index = -1 // for safety
	t.cur--
	//log.Debug("pop cur: %d", t.cur)
	return
}

func (t *Timer) process() {
	var d time.Duration
	log.Debug("strat timer process")
	for {
		d = t.Find()
		if d > zeroDuration {
			if d > maxTimerDelay {
				d = maxTimerDelay
			}
		} else if d == infiniteDuration {
			d = timerDelay
		} else {
			t.Expire()
			continue
		}
		time.Sleep(d)
	}
}

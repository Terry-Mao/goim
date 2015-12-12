package main

import (
	log "code.google.com/p/log4go"
	"io"
	"sync"
	"time"
)

const (
	timerFormat      = "2006-01-02 15:04:05"
	zeroDuration     = time.Duration(0)
	infiniteDuration = time.Duration(-1)
	timerDelay       = 100 * time.Millisecond
	maxTimerDelay    = 500 * time.Millisecond
	timerLazyDelay   = 300 * time.Millisecond
)

type TimerData struct {
	Key    string
	Expire time.Time
	Value  io.Closer
	index  int
	next   *TimerData
}

func (td *TimerData) Delay() time.Duration {
	return td.Expire.Sub(time.Now())
}

func (td *TimerData) Lazy(expire time.Duration) bool {
	return time.Now().Add(expire).Sub(td.Expire) < timerLazyDelay
}

func (td *TimerData) ExpireString() string {
	return td.Expire.Format(timerFormat)
}

type Timer struct {
	lock   sync.Mutex
	free   *TimerData
	timers []*TimerData
	cur    int
}

// get get a free timer data.
func (t *Timer) get() (td *TimerData, err error) {
	td = t.free
	t.free = td.next
	if td == nil {
		err = ErrTimerFull
	}
	return
}

// put put back a timer data.
func (t *Timer) put(td *TimerData) {
	td.next = t.free
	t.free = td
}

// A heap must be initialized before any of the heap operations
// can be used. Init is idempotent with respect to the heap invariants
// and may be called whenever the heap invariants may have been invalidated.
// Its complexity is O(n) where n = h.Len().
//
func NewTimer(num int) (t *Timer) {
	t = new(Timer)
	t.init(num)
	return t
}

func (t *Timer) init(num int) {
	t.timers = make([]*TimerData, num)
	t.cur = -1
	tds := make([]TimerData, num)
	t.free = &(tds[0])
	td := t.free
	for i := 1; i < num; i++ {
		td.next = &(tds[i])
		td = td.next
	}
}

// Init init the timer.
func (t *Timer) Init(num int) {
	t.init(num)
}

// Push pushes the element x onto the heap. The complexity is
// O(log(n)) where n = h.Len().
func (t *Timer) Add(expire time.Duration, closer io.Closer) (td *TimerData, err error) {
	t.lock.Lock()
	if td, err = t.get(); err == nil {
		td.Expire = time.Now().Add(expire)
		td.Value = closer
		t.add(td)
	}
	t.lock.Unlock()
	return
}

// Push pushes the element x onto the heap. The complexity is
// O(log(n)) where n = h.Len().
func (t *Timer) add(td *TimerData) {
	t.cur++
	td.index = t.cur
	// add to the minheap last node
	t.timers[t.cur] = td
	t.up(t.cur)
	if Debug {
		log.Debug("timer: push item key: %s, expire: %s, index: %d", td.Key, td.ExpireString(), td.index)
	}
	return
}

// Del removes the element at index i from the heap.
// The complexity is O(log(n)) where n = h.Len().
func (t *Timer) Del(td *TimerData) {
	t.lock.Lock()
	if td.index != -1 {
		// already remove, usually by expire
		t.del(td.index)
	}
	t.put(td)
	t.lock.Unlock()
	if Debug {
		log.Debug("timer: remove item key: %s, expire: %s, index: %d", td.Key, td.ExpireString(), td.index)
	}
	return
}

func (t *Timer) del(i int) {
	if t.cur != i {
		t.swap(i, t.cur)
		t.down(i, t.cur)
		t.up(i)
	}
	// remove item is the last node
	t.timers[t.cur].index = -1 // for safety
	t.cur--
}

// Set update timer data.
func (t *Timer) Set(td *TimerData, expire time.Duration) {
	t.lock.Lock()
	if td.index != -1 {
		// already remove, usually by expire
		t.del(td.index)
	}
	td.Expire = time.Now().Add(expire)
	t.add(td)
	t.lock.Unlock()
	return
}

// Find a expire timer data, if not found ,return infinite time.
func (t *Timer) Find() (d time.Duration) {
	t.lock.Lock()
	if t.cur < 0 {
		d = infiniteDuration
	} else {
		d = t.timers[0].Delay()
	}
	t.lock.Unlock()
	return
}

// Expire removes the minimum element (according to Less) from the heap.
// The complexity is O(log(n)) where n = max.
// It is equivalent to Del(0).
func (t *Timer) Expire() {
	var (
		err error
		d   time.Duration
		td  *TimerData
	)
	t.lock.Lock()
	for t.cur >= 0 {
		td = t.timers[0]
		if d = td.Delay(); d > 0 {
			break
		}
		if Debug {
			log.Debug("find a expire timer key: %s, expire: %s, index: %d", td.Key, td.ExpireString(), td.index)
		}
		if td.Value == nil {
			log.Warn("expire timer no io.Closer")
		} else {
			if Debug {
				log.Debug("timer key: %s, expire: %s, index: %d expired, call Close()", td.Key, td.ExpireString, td.index)
			}
			if err = td.Value.Close(); err != nil {
				log.Error("timer close error(%v)", err)
			}
		}
		t.del(0)
	}
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
	return t.timers[i].Expire.Before(t.timers[j].Expire)
}

func (t *Timer) swap(i, j int) {
	//log.Debug("swap(%d, %d)", i, j)
	t.timers[i], t.timers[j] = t.timers[j], t.timers[i]
	t.timers[i].index = i
	t.timers[j].index = j
}

// TimerProcess one process goroutine handle many timers.
// range all timers call find the time.Duration
// sleep
// range all timers call expire
func TimerProcess(timers []Timer) {
	var (
		i  int
		t  *Timer
		d  time.Duration
		md = timerDelay
	)
	// loop forever
	for {
		for i = 0; i < len(timers); i++ {
			t = &(timers[i])
			d = t.Find()
			if d > zeroDuration {
				if d > maxTimerDelay {
					d = maxTimerDelay
				}
			} else if d == infiniteDuration {
				d = timerDelay
			} else {
				// if found call expire and calculate again the min timerd
				t.Expire()
				d = zeroDuration
				break
			}
			if md > d {
				md = d
			}
		}
		if d != zeroDuration {
			if Debug {
				log.Debug("timer process sleep: %s", md.String())
			}
			time.Sleep(md)
			md = timerDelay
		}
	}
}

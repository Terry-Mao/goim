package main

import (
	"sync"
	"time"
)

const (
	maxCleanNum = 100
)

type CleanData struct {
	Key        int64
	expireTime time.Time
	next, prev *CleanData
}

type Cleaner struct {
	cLock sync.Mutex
	size  int
	root  CleanData
	maps  map[int64]*CleanData
}

func NewCleaner(cleaner int) *Cleaner {
	c := new(Cleaner)
	c.root.next = &c.root
	c.root.prev = &c.root
	c.size = 0
	c.maps = make(map[int64]*CleanData, cleaner)
	return c
}

func (c *Cleaner) PushFront(key int64, expire time.Duration) {
	c.cLock.Lock()
	if e, ok := c.maps[key]; ok {
		// update time
		e.expireTime = time.Now().Add(expire)
		c.moveToFront(e)
	} else {
		e = new(CleanData)
		e.Key = key
		e.expireTime = time.Now().Add(expire)
		c.maps[key] = e
		at := &c.root
		n := at.next
		at.next = e
		e.prev = at
		e.next = n
		n.prev = e
		c.size++
	}
	c.cLock.Unlock()
}

func (c *CleanData) expire() bool {
	return c.expireTime.Before(time.Now())
}

func (c *Cleaner) moveToFront(e *CleanData) {
	if c.root.next != e {
		at := &c.root
		// remove element
		e.prev.next = e.next
		e.next.prev = e.prev
		n := at.next
		at.next = e
		e.prev = at
		e.next = n
		n.prev = e
	}
}

func (c *Cleaner) Remove(key int64) {
	c.cLock.Lock()
	c.remove(key)
	c.cLock.Unlock()
}

func (c *Cleaner) remove(key int64) {
	if e, ok := c.maps[key]; ok {
		delete(c.maps, key)
		e.prev.next = e.next
		e.next.prev = e.prev
		e.next = nil // avoid memory leaks
		e.prev = nil // avoid memory leaks
		c.size--
	}
}

func (c *Cleaner) Clean() (keys []int64) {
	var (
		i int
		e *CleanData
	)
	keys = make([]int64, 0, maxCleanNum)
	c.cLock.Lock()
	for i = 0; i < maxCleanNum; i++ {
		if e = c.back(); e != nil {
			if e.expire() {
				c.remove(e.Key)
				keys = append(keys, e.Key)
				continue
			}
		}
		break
	}
	// next time
	c.cLock.Unlock()
	return
}

func (c *Cleaner) back() *CleanData {
	if c.size == 0 {
		return nil
	}
	return c.root.prev
}

package main

import (
	"testing"
	"time"
)

func TestCleaner(t *testing.T) {
	c := NewCleaner(10)
	c.PushFront(1, time.Second*1)
	time.Sleep(3 * time.Second)
	keys := c.Clean()
	if len(keys) == 0 {
		t.FailNow()
	}
}

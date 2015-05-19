package main

import (
	//	log "code.google.com/p/log4go"
	"sync"
)

type Node struct {
	server int16
	sLock  sync.Mutex
	subs   map[string]int8
}

// Bucket is a channel holder.
type Bucket struct {
	tLock  sync.Mutex
	topics map[string][]Node
}

package main

type Bucket struct {
	lock sync.Mutex
}

package main

import (
	log "code.google.com/p/log4go"
)

const (
	// signal command
	SignalNum   = 1
	ProtoFinish = 0
	ProtoReady  = 1
)

type Ring struct {
	// read
	// rn int
	rp int
	// add data here split reade & write in one cacheline
	data []Proto
	num  int
	// write
	// wn int
	wp int
}

func NewRing(num int) *Ring {
	r := new(Ring)
	r.init(num)
	return r
}

func (r *Ring) init(num int) {
	r.data = make([]Proto, num)
	r.num = num
}

func (r *Ring) Init(num int) {
	r.init(num)
}

func (r *Ring) Get() (proto *Proto, err error) {
	if r.rp == r.wp {
		return nil, ErrRingEmpty
	}
	proto = &r.data[r.rp%r.num]
	return
}

func (r *Ring) GetAdv() {
	r.rp++
	if Debug {
		log.Debug("ring rp: %d, idx: %d", r.rp, r.rp%r.num)
	}
}

func (r *Ring) Set() (proto *Proto, err error) {
	if r.wp-r.rp >= r.num {
		return nil, ErrRingFull
	}
	proto = &r.data[r.wp%r.num]
	return
}

func (r *Ring) SetAdv() {
	r.wp++
	if Debug {
		log.Debug("ring wp: %d, idx: %d", r.wp, r.wp%r.num)
	}
}

func (r *Ring) Reset() {
	r.rp = 0
	r.wp = 0
}

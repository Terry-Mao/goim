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
	mask int
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
	num = (num + 1) & ^1
	r.data = make([]Proto, num)
	r.num = num
	r.mask = r.num - 1
}

func (r *Ring) Init(num int) {
	r.init(num)
}

func (r *Ring) Get() (proto *Proto, err error) {
	if r.rp == r.wp {
		return nil, ErrRingEmpty
	}
	proto = &r.data[r.rp&r.mask]
	return
}

func (r *Ring) GetAdv() {
	r.rp++
	if Debug {
		log.Debug("ring rp: %d, idx: %d", r.rp, r.rp&r.mask)
	}
}

func (r *Ring) Set() (proto *Proto, err error) {
	if r.wp-r.rp >= r.num {
		return nil, ErrRingFull
	}
	proto = &r.data[r.wp&r.mask]
	return
}

func (r *Ring) SetAdv() {
	r.wp++
	if Debug {
		log.Debug("ring wp: %d, idx: %d", r.wp, r.wp&r.mask)
	}
}

func (r *Ring) Reset() {
	r.rp = 0
	r.wp = 0
}

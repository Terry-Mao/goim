package main

import (
	log "code.google.com/p/log4go"
)

type Ring struct {
	// read
	rn int
	rp int
	// write
	wn int
	wp int
	// info
	num  int
	data []Proto
}

func NewRing(num int) *Ring {
	r := new(Ring)
	r.data = make([]Proto, num)
	r.num = num
	return r
}

func InitRing(r *Ring, num int) {
	r.data = make([]Proto, num)
	r.num = num
}

func (r *Ring) Get() (proto *Proto, err error) {
	if r.wn == r.rn {
		return nil, ErrRingEmpty
	}
	proto = &r.data[r.rp]
	return
}

func (r *Ring) GetAdv() {
	if r.rp++; r.rp >= r.num {
		r.rp = 0
	}
	r.rn++
	log.Debug("ring rn: %d, rp: %d", r.rn, r.rp)
}

func (r *Ring) Set() (proto *Proto, err error) {
	if r.wn-r.rn >= r.num {
		return nil, ErrRingFull
	}
	proto = &r.data[r.wp]
	return
}

func (r *Ring) SetAdv() {
	if r.wp++; r.wp >= r.num {
		r.wp = 0
	}
	r.wn++
	log.Debug("ring wn: %d, wp: %d", r.wn, r.wp)
}

func (r *Ring) Reset() {
	r.rn = 0
	r.rp = 0
	r.wn = 0
	r.wp = 0
}

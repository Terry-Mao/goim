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
	rn int
	
	// write
	pad [64]byte
	
	wn int
	
	// info
	num  int
	data []Proto
}

func NewRing(num int) *Ring {
	r := new(Ring)
	r.init(initNum(num))
	return r
}

func initNum(number int) (ret int) {
	if (number & (number-1) == 0) {
		return number
	} else {
		for (number & (number-1) != 0){
			 number &= (number-1)
			 }
		return number << 1;
	}
}

func (r *Ring) init(num int) {
	r.data = make([]Proto, num)
	r.num = num
}

func (r *Ring) Init(num int) {
	r.init(num)
}

func (r *Ring) Get() (proto *Proto, err error) {
	if r.wn == r.rn {
		return nil, ErrRingEmpty
	}
	proto = &r.data[r.rn&(r.num-1)]
	return
}

func (r *Ring) GetAdv() {
	if r.rn++; r.rn >= maxInt {
		r.rn = 0
	}
	r.rn++
	if Debug {
		log.Debug("ring rn: %d, num: %d", r.rn, r.num)
	}
}

func (r *Ring) Set() (proto *Proto, err error) {
	if r.wn-r.rn >= r.num {
		return nil, ErrRingFull
	}
	proto = &r.data[r.wn&(r.num-1)]
	return
}

func (r *Ring) SetAdv() {
	if r.wn++; r.wn >= maxInt {
		r.wn = 0
	}
	r.wn++
	if Debug {
		log.Debug("ring wn: %d, num: %d", r.wn, r.num)
	}
}

func (r *Ring) Reset() {
	r.rn = 0
	r.wn = 0
}

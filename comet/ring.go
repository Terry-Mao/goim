package main

import (
	log "code.google.com/p/log4go"
)

const (
	// signal command
	SignalNum   = 1
	ProtoFinish = 0
	ProtoReady  = 1
	maxUint64  = ^uint64(0) 
)

type Ring struct {
	// read
	rn uint64
	
	// info
	num  uint64
	numMask  uint64
	pad [40]byte
	
	// write
	wn uint64
	
	
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
	r.num = uint64(num)
	r.numMask = r.num-1
}

func (r *Ring) Init(num int) {
	r.init(num)
}

func (r *Ring) Get() (proto *Proto, err error) {
	if r.wn == r.rn {
		return nil, ErrRingEmpty
	}
	proto = &r.data[r.rn&r.numMask]
	return
}

func (r *Ring) GetAdv() {
	if r.rn++; r.rn >= maxUint64 {
		r.rn = 0
	}
	if Debug {
		log.Debug("ring rn: %d, num: %d", r.rn, r.num)
	}
}

func (r *Ring) Set() (proto *Proto, err error) {
	if r.wn-r.rn >= r.num {
		return nil, ErrRingFull
	}
	proto = &r.data[r.wn&r.numMask]
	return
}

func (r *Ring) SetAdv() {
	if r.wn++; r.wn >= maxUint64 {
		r.wn = 0
	}
	if Debug {
		log.Debug("ring wn: %d, num: %d", r.wn, r.num)
	}
}

func (r *Ring) Reset() {
	r.rn = 0
	r.wn = 0
}

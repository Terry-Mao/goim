package main

import (
	log "code.google.com/p/log4go"
	"crypto/cipher"
	"github.com/Terry-Mao/goim/libs/crypto/aes"
	"sync"
)

// Proto is a request&response written before every goim connect.  It is used internally
// but documented here as an aid to debugging, such as when analyzing
// network traffic.
type Proto struct {
	PackLen   int32  // package length
	HeaderLen int16  // header length
	Ver       int16  // protocol version
	Operation int32  // operation for request
	SeqId     int32  // sequence number chosen by client
	Body      []byte // body
	next      *Proto // free list
}

func (p *Proto) Encrypt(block cipher.Block) (err error) {
	if p.Body != nil {
		p.Body, err = aes.ECBEncrypt(block, p.Body)
	}
	return
}

func (p *Proto) Decrypt(block cipher.Block) (err error) {
	if p.Body != nil {
		p.Body, err = aes.ECBDecrypt(block, p.Body)
	}
	return
}

type FreeProto struct {
	free *Proto     // free pointer
	lock sync.Mutex // protected protos
	used int        // protos used
}

func NewFreeProto(num int) *FreeProto {
	f := new(FreeProto)
	p := new(Proto)
	f.free = p
	for i := 1; i < num; i++ {
		p.next = new(Proto)
		p = p.next
	}
	return f
}

func (f *FreeProto) Get() (p *Proto) {
	f.lock.Lock()
	p = f.free
	if p != nil {
		f.free = p.next
		f.used++
		f.lock.Unlock()
		log.Debug("get proto, used: %d", f.used)
	} else {
		f.lock.Unlock()
		p = new(Proto)
	}
	return
}

func (f *FreeProto) Free(p *Proto) {
	f.lock.Lock()
	// if no used proto, free list full, discard it
	if f.used == 0 {
		f.lock.Unlock()
		// use gc free
		return
	}
	p.next = f.free
	f.free = p
	f.used--
	f.lock.Unlock()
	log.Debug("put timerdata, used: %d", f.used)
}

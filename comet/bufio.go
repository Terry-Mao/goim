package main

import (
	"bufio"
	"io"
	"sync"
)

// TODO
func NewBufioReaderSize(pool *sync.Pool, r io.Reader, size int) *bufio.Reader {
	if v := pool.Get(); v != nil {
		br := v.(*bufio.Reader)
		br.Reset(r)
		return br
	}
	return bufio.NewReaderSize(r, size)
}

func PutBufioReader(pool *sync.Pool, br *bufio.Reader) {
	br.Reset(nil)
	pool.Put(br)
}

func NewBufioWriterSize(pool *sync.Pool, w io.Writer, size int) *bufio.Writer {
	if v := pool.Get(); v != nil {
		bw := v.(*bufio.Writer)
		bw.Reset(w)
		return bw
	}
	return bufio.NewWriterSize(w, size)
}

func PutBufioWriter(pool *sync.Pool, bw *bufio.Writer) {
	bw.Reset(nil)
	pool.Put(bw)
}

func NewByteArraySize(pool *sync.Pool, size int) []byte {
	if v := pool.Get(); v != nil {
		ba := v.([]byte)
		return ba
	}
	return make([]byte, size)
}

func PutByteArray(pool *sync.Pool, ba []byte) {
	pool.Put(ba)
}

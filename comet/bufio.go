package main

import (
	"bufio"
	"io"
	"sync"
)

func newBufioReaderSize(pool *sync.Pool, r io.Reader, size int) *bufio.Reader {
	if v := pool.Get(); v != nil {
		br := v.(*bufio.Reader)
		br.Reset(r)
		return br
	}
	return bufio.NewReaderSize(r, size)
}

func putBufioReader(pool *sync.Pool, br *bufio.Reader) {
	br.Reset(nil)
	pool.Put(br)
}

func newBufioWriterSize(pool *sync.Pool, w io.Writer, size int) *bufio.Writer {
	if v := pool.Get(); v != nil {
		bw := v.(*bufio.Writer)
		bw.Reset(w)
		return bw
	}
	return bufio.NewWriterSize(w, size)
}

func putBufioWriter(pool *sync.Pool, bw *bufio.Writer) {
	bw.Reset(nil)
	pool.Put(bw)
}

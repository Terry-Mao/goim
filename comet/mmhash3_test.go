// Copyright (c) 2012 The Go Authors. All rights reserved.
// refer https://github.com/gwenn/murmurhash3

package main

import (
	"hash"
	"testing"
)

const testDataSize = 40

func TestMurmur3A(t *testing.T) {
	expected := uint32(3127628307)
	hash := Murmur3A([]byte("test"), 0)
	if hash != expected {
		t.Errorf("Expected %d but was %d for Murmur3A\n", expected, hash)
	}
}

func TestMurmur3C(t *testing.T) {
	expected := []uint32{1862463280, 1426881896, 1426881896, 1426881896}
	hash := Murmur3C([]byte("test"), 0)
	for i, e := range expected {
		if hash[i] != e {
			t.Errorf("Expected %d but was %d for Murmur3C[%d]\n", e, hash[i], i)
		}
	}
}

func TestMurmur3F(t *testing.T) {
	expected := []uint64{12429135405209477533, 11102079182576635266}
	hash := Murmur3F([]byte("test"), 0)
	for i, e := range expected {
		if hash[i] != e {
			t.Errorf("Expected %d but was %d for Murmur3F[%d]\n", e, hash[i], i)
		}
	}
}

func Benchmark3A(b *testing.B) {
	benchmark(b, NewMurmur3A())
}
func Benchmark3C(b *testing.B) {
	benchmark(b, NewMurmur3C())
}
func Benchmark3F(b *testing.B) {
	benchmark(b, NewMurmur3F())
}

func benchmark(b *testing.B, h hash.Hash) {
	b.ResetTimer()
	b.SetBytes(testDataSize)
	data := make([]byte, testDataSize)
	for i := range data {
		data[i] = byte(i + 'a')
	}

	b.StartTimer()
	for todo := b.N; todo != 0; todo-- {
		h.Reset()
		h.Write(data)
		h.Sum(nil)
	}
}

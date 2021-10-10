package bytes

import (
	"fmt"
	"reflect"
	"testing"
)

func TestWriter(t *testing.T) {
	w := NewWriterSize(64)
	if w.Len() != 0 && w.Size() != 64 {
		t.FailNow()
	}
	b := []byte("hello")
	w.Write(b)
	if !reflect.DeepEqual(b, w.Buffer()) {
		t.FailNow()
	}
	w.Peek(len(b))
	w.Reset()
	for i := 0; i < 1024; i++ {
		w.Write(b)
	}
	w.Reset()
	if w.Len() != 0 {
		t.FailNow()
	}
}

/*
=== RUN   TestWriter2
len 0 size 64 n 0
buf: [0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0]
len 2 size 64 n 2
buf: [1 2 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0]
peekBs: [0 0]
len 4 size 64 n 4
buf: [1 2 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0]
len 0 size 64 n 0
buf: [1 2 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0]
*/
func TestWriter2(t *testing.T) {
	w := NewWriterSize(64)
	if w.Len() != 0 && w.Size() != 64 {
		t.FailNow()
	}
	w.print()

	b := []byte{byte(1),byte(2)}
	w.Write(b)
	w.print()

	peekBs := w.Peek(len(b))
	fmt.Printf("peekBs: %v\n", peekBs)
	w.print()

	w.Reset()
	w.print()
}

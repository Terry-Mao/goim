package bytes

import (
	"fmt"
	"testing"
	"unsafe"
)

func TestBuffer(t *testing.T) {
	var v unsafe.Pointer
	num := 2
	sz := 10
	p := NewPool(num, sz)
	b := p.Get()

	if b.Bytes() == nil || len(b.Bytes()) == 0 {
		t.FailNow()
	}
	v = unsafe.Pointer(&b.buf[0])
	// 0xc000016300
	fmt.Printf("v %v\n", v)
	b = p.Get()
	if b.Bytes() == nil || len(b.Bytes()) == 0 {
		t.FailNow()
	}
	v1 := unsafe.Pointer(&b.buf[0])
	// 0xc00001630a 减去上个v是 size大小
	fmt.Printf("v1 %v\n", v1)
	// num > 1的情况下
	//if int(*(int*)(v1)-*v) != sz {
	//	t.FailNow()
	//}

	b = p.Get()
	if b.Bytes() == nil || len(b.Bytes()) == 0 {
		t.FailNow()
	}
	v2 := unsafe.Pointer(&b.buf[0])
	// 0xc000016320
	fmt.Printf("v2 %v\n", v2)
}

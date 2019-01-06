package binary

import "testing"

func TestInt8(t *testing.T) {
	b := make([]byte, 1)
	BigEndian.PutInt8(b, 100)
	i := BigEndian.Int8(b)
	if i != 100 {
		t.FailNow()
	}
}

func TestInt16(t *testing.T) {
	b := make([]byte, 2)
	BigEndian.PutInt16(b, 100)
	i := BigEndian.Int16(b)
	if i != 100 {
		t.FailNow()
	}
}

func TestInt32(t *testing.T) {
	b := make([]byte, 4)
	BigEndian.PutInt32(b, 100)
	i := BigEndian.Int32(b)
	if i != 100 {
		t.FailNow()
	}
}

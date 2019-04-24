package strings

import (
	"reflect"
	"testing"
)

func TestInt32(t *testing.T) {
	i := []int32{1, 2, 3}
	s := JoinInt32s(i, ",")
	ii, _ := SplitInt32s(s, ",")
	if !reflect.DeepEqual(i, ii) {
		t.FailNow()
	}
}

func TestInt64(t *testing.T) {
	i := []int64{1, 2, 3}
	s := JoinInt64s(i, ",")
	ii, _ := SplitInt64s(s, ",")
	if !reflect.DeepEqual(i, ii) {
		t.FailNow()
	}
}

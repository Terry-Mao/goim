package strings

import (
	"bytes"
	"strconv"
	"strings"
	"sync"
)

var (
	bfPool = sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer([]byte{})
		},
	}
)

// JoinInt32s format int323232 slice like:n1,n2,n3.
func JoinInt32s(is []int32, p string) string {
	if len(is) == 0 {
		return ""
	}
	if len(is) == 1 {
		return strconv.FormatInt(int64(is[0]), 10)
	}
	buf := bfPool.Get().(*bytes.Buffer)
	for _, i := range is {
		buf.WriteString(strconv.FormatInt(int64(i), 10))
		buf.WriteString(p)
	}
	if buf.Len() > 0 {
		buf.Truncate(buf.Len() - 1)
	}
	s := buf.String()
	buf.Reset()
	bfPool.Put(buf)
	return s
}

// SplitInt32s split string into int32 slice.
func SplitInt32s(s, p string) ([]int32, error) {
	if s == "" {
		return nil, nil
	}
	sArr := strings.Split(s, p)
	res := make([]int32, 0, len(sArr))
	for _, sc := range sArr {
		i, err := strconv.ParseInt(sc, 10, 32)
		if err != nil {
			return nil, err
		}
		res = append(res, int32(i))
	}
	return res, nil
}

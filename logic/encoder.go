package main

import (
	"fmt"
	"strconv"
	"strings"
)

func encode(userId int64, seq int32) string {
	return fmt.Sprintf("%d_%d", userId, seq)
}

func decode(key string) (userId int64, seq int32, err error) {
	var (
		idx int
		t   int64
	)
	if idx = strings.IndexByte(key, '_'); idx == -1 {
		err = ErrDecodeKey
		return
	}
	if userId, err = strconv.ParseInt(key[:idx], 10, 64); err != nil {
		return
	}
	if t, err = strconv.ParseInt(key[idx+1:], 10, 32); err != nil {
		return
	}
	seq = int32(t)
	return
}

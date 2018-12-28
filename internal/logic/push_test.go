package logic

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPushKeys(t *testing.T) {
	var (
		c    = context.TODO()
		op   = int32(100)
		keys = []string{"test_key"}
		msg  = []byte("hello")
	)
	err := lg.PushKeys(c, op, keys, msg)
	assert.Nil(t, err)
}

func TestPushMids(t *testing.T) {
	var (
		c    = context.TODO()
		op   = int32(100)
		mids = []int64{1, 2, 3}
		msg  = []byte("hello")
	)
	err := lg.PushMids(c, op, mids, msg)
	assert.Nil(t, err)
}

func TestPushRoom(t *testing.T) {
	var (
		c    = context.TODO()
		op   = int32(100)
		typ  = "test"
		room = "test_room"
		msg  = []byte("hello")
	)
	err := lg.PushRoom(c, op, typ, room, msg)
	assert.Nil(t, err)
}

func TestPushAll(t *testing.T) {
	var (
		c     = context.TODO()
		op    = int32(100)
		speed = int32(100)
		msg   = []byte("hello")
	)
	err := lg.PushAll(c, op, speed, msg)
	assert.Nil(t, err)
}

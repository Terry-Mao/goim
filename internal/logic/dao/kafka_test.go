package dao

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDaoPushMsg(t *testing.T) {
	var (
		c      = context.Background()
		op     = int32(100)
		server = "test"
		msg    = []byte("msg")
		keys   = []string{"key"}
	)
	err := d.PushMsg(c, op, server, keys, msg)
	assert.Nil(t, err)
}

func TestDaoBroadcastRoomMsg(t *testing.T) {
	var (
		c    = context.Background()
		op   = int32(100)
		room = "test://1"
		msg  = []byte("msg")
	)
	err := d.BroadcastRoomMsg(c, op, room, msg)
	assert.Nil(t, err)
}

func TestDaoBroadcastMsg(t *testing.T) {
	var (
		c     = context.Background()
		op    = int32(100)
		speed = int32(0)
		msg   = []byte("")
	)
	err := d.BroadcastMsg(c, op, speed, msg)
	assert.Nil(t, err)
}

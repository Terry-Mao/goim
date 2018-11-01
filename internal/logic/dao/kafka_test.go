package dao

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDaoPushMsg(t *testing.T) {
	var (
		c      = context.Background()
		op     = int32(0)
		server = ""
		msg    = ""
		keys   = []string{"key"}
	)
	err := d.PushMsg(c, op, server, msg, keys)
	assert.Nil(t, err)
}

func TestDaoBroadcastRoomMsg(t *testing.T) {
	var (
		c    = context.Background()
		op   = int32(0)
		room = ""
		msg  = ""
	)
	err := d.BroadcastRoomMsg(c, op, room, msg)
	assert.Nil(t, err)
}

func TestDaoBroadcastMsg(t *testing.T) {
	var (
		c        = context.Background()
		op       = int32(0)
		speed    = int32(0)
		msg      = ""
		platform = ""
	)
	err := d.BroadcastMsg(c, op, speed, msg, platform)
	assert.Nil(t, err)
}

package logic

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOnline(t *testing.T) {
	var (
		c     = context.TODO()
		typ   = "test"
		n     = 2
		rooms = []string{"room_01", "room_02", "room_03"}
	)
	lg.totalIPs = 100
	lg.totalConns = 200
	lg.roomCount = map[string]int32{
		"test://room_01": 100,
		"test://room_02": 200,
		"test://room_03": 300,
	}
	tops, err := lg.OnlineTop(c, typ, n)
	assert.Nil(t, err)
	assert.Equal(t, len(tops), 2)
	onlines, err := lg.OnlineRoom(c, typ, rooms)
	assert.Nil(t, err)
	assert.Equal(t, onlines["room_01"], int32(100))
	assert.Equal(t, onlines["room_02"], int32(200))
	assert.Equal(t, onlines["room_03"], int32(300))
	ips, conns := lg.OnlineTotal(c)
	assert.Equal(t, ips, int64(100))
	assert.Equal(t, conns, int64(200))
}

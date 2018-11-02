package logic

import (
	"context"
	"testing"

	"github.com/issue9/assert"
)

func TestConnect(t *testing.T) {
	var (
		server    = "test_server"
		serverKey = "test_server_key"
		token     = []byte(`1|test_server_key|live://test_room|web|1000,1001,1002`)
		c         = context.Background()
	)
	// connect
	mid, key, roomID, _, accepts, err := l.Connect(c, server, serverKey, "", token)
	assert.Nil(t, err)
	assert.Equal(t, serverKey, key)
	assert.Equal(t, roomID, "live://test_room")
	assert.Equal(t, len(accepts), 3)
	t.Log(mid, key, roomID, accepts, err)
	// heartbeat
	err = l.Heartbeat(c, mid, key, server)
	assert.Nil(t, err)
	// disconnect
	has, err := l.Disconnect(c, mid, key, server)
	assert.Nil(t, err)
	assert.Equal(t, true, has)
}

package dao

import (
	"context"
	"testing"

	"github.com/Terry-Mao/goim/internal/logic/model"
	"github.com/stretchr/testify/assert"
)

func TestDaopingRedis(t *testing.T) {
	err := d.pingRedis(context.Background())
	assert.Nil(t, err)
}

func TestDaoAddMapping(t *testing.T) {
	var (
		c      = context.Background()
		mid    = int64(1)
		key    = "test_key"
		server = "test_server"
	)
	err := d.AddMapping(c, 0, "test", server)
	assert.Nil(t, err)
	err = d.AddMapping(c, mid, key, server)
	assert.Nil(t, err)

	has, err := d.ExpireMapping(c, 0, "test")
	assert.Nil(t, err)
	assert.NotEqual(t, false, has)
	has, err = d.ExpireMapping(c, mid, key)
	assert.Nil(t, err)
	assert.NotEqual(t, false, has)

	res, err := d.ServersByKeys(c, []string{key})
	assert.Nil(t, err)
	assert.Equal(t, server, res[0])

	ress, mids, err := d.KeysByMids(c, []int64{mid})
	assert.Nil(t, err)
	assert.Equal(t, server, ress[key])
	assert.Equal(t, mid, mids[0])

	has, err = d.DelMapping(c, 0, "test", server)
	assert.Nil(t, err)
	assert.NotEqual(t, false, has)
	has, err = d.DelMapping(c, mid, key, server)
	assert.Nil(t, err)
	assert.NotEqual(t, false, has)
}

func TestDaoAddServerOnline(t *testing.T) {
	var (
		c      = context.Background()
		server = "test_server"
		online = &model.Online{
			RoomCount: map[string]int32{"room": 10},
		}
	)
	err := d.AddServerOnline(c, server, online)
	assert.Nil(t, err)

	r, err := d.ServerOnline(c, server)
	assert.Nil(t, err)
	assert.Equal(t, online.RoomCount["room"], r.RoomCount["room"])

	err = d.DelServerOnline(c, server)
	assert.Nil(t, err)
}

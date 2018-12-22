package logic

import (
	"context"
	"testing"

	"github.com/Terry-Mao/goim/internal/logic/model"

	"github.com/stretchr/testify/assert"
)

func TestNodes(t *testing.T) {
	var (
		c        = context.TODO()
		clientIP = "127.0.0.1"
	)
	ins, err := lg.NodesInstances(c)
	assert.Nil(t, err)
	assert.NotNil(t, ins)
	nodes := lg.NodesWeighted(c, model.PlatformWeb, clientIP)
	assert.NotNil(t, nodes)
	nodes = lg.NodesWeighted(c, "android", clientIP)
	assert.NotNil(t, nodes)
}

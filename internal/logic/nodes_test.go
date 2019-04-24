package logic

import (
	"context"
	"testing"

	"github.com/bilibili/discovery/naming"
	"github.com/Terry-Mao/goim/internal/logic/model"

	"github.com/stretchr/testify/assert"
)

func TestNodes(t *testing.T) {
	var (
		c        = context.TODO()
		clientIP = "127.0.0.1"
	)
	lg.nodes = make([]*naming.Instance, 0)
	ins := lg.NodesInstances(c)
	assert.NotNil(t, ins)
	nodes := lg.NodesWeighted(c, model.PlatformWeb, clientIP)
	assert.NotNil(t, nodes)
	nodes = lg.NodesWeighted(c, "android", clientIP)
	assert.NotNil(t, nodes)
}

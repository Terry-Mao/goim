package http

import (
	"context"

	"github.com/gin-gonic/gin"
)

func (s *Server) nodesWeighted(c *gin.Context) {
	var arg struct {
		Platform string `form:"platform"`
	}
	if err := c.BindQuery(&arg); err != nil {
		result(c, nil, RequestErr)
		return
	}
	res := s.logic.NodesWeighted(c, arg.Platform, c.ClientIP())
	result(c, res, OK)
}

func (s *Server) nodesInstances(c *gin.Context) {
	res, err := s.logic.NodesInstances(context.TODO())
	if err != nil {
		result(c, nil, ServerErr)
		return
	}
	result(c, res, OK)
}

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
		errors(c, RequestErr, err.Error())
		return
	}
	res := s.logic.NodesWeighted(c, arg.Platform, c.ClientIP())
	result(c, res, OK)
}

func (s *Server) nodesInstances(c *gin.Context) {
	res := s.logic.NodesInstances(context.TODO())
	result(c, res, OK)
}

package http

import (
	"context"

	"github.com/gin-gonic/gin"
)

func (s *Server) nodesWeighted(c *gin.Context) {
	var arg struct {
		Platform string `form:"platform"`
	}
	if err := c.Bind(arg); err != nil {
		writeJSON(c, nil, RequestErr)
		return
	}
	res := s.logic.NodesWeighted(c, arg.Platform, c.ClientIP())
	writeJSON(c, res, OK)
}

func (s *Server) nodesInstances(c *gin.Context) {
	res, err := s.logic.NodesInstances(context.TODO())
	if err != nil {
		writeJSON(c, nil, ServerErr)
		return
	}
	writeJSON(c, res, OK)
}

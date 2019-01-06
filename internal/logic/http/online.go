package http

import (
	"context"

	"github.com/gin-gonic/gin"
)

func (s *Server) onlineTop(c *gin.Context) {
	var arg struct {
		Type  string `form:"type" binding:"required"`
		Limit int    `form:"limit" binding:"required"`
	}
	if err := c.BindQuery(&arg); err != nil {
		errors(c, RequestErr, err.Error())
		return
	}
	res, err := s.logic.OnlineTop(c, arg.Type, arg.Limit)
	if err != nil {
		result(c, nil, RequestErr)
		return
	}
	result(c, res, OK)
}

func (s *Server) onlineRoom(c *gin.Context) {
	var arg struct {
		Type  string   `form:"type" binding:"required"`
		Rooms []string `form:"rooms" binding:"required"`
	}
	if err := c.BindQuery(&arg); err != nil {
		errors(c, RequestErr, err.Error())
		return
	}
	res, err := s.logic.OnlineRoom(c, arg.Type, arg.Rooms)
	if err != nil {
		result(c, nil, RequestErr)
		return
	}
	result(c, res, OK)
}

func (s *Server) onlineTotal(c *gin.Context) {
	ipCount, connCount := s.logic.OnlineTotal(context.TODO())
	res := map[string]interface{}{
		"ip_count":   ipCount,
		"conn_count": connCount,
	}
	result(c, res, OK)
}

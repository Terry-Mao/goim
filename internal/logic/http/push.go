package http

import (
	"context"
	"io/ioutil"

	"github.com/gin-gonic/gin"
)

func (s *Server) pushKeys(c *gin.Context) {
	var arg struct {
		Op   int32    `form:"operation"`
		Keys []string `form:"keys"`
	}
	if err := c.Bind(arg); err != nil {
		writeJSON(c, nil, RequestErr)
		return
	}
	// read message
	msg, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		writeJSON(c, nil, RequestErr)
		return
	}
	if err = s.logic.PushKeys(context.TODO(), arg.Op, arg.Keys, msg); err != nil {
		writeJSON(c, nil, RequestErr)
		return
	}
	writeJSON(c, nil, OK)
}

func (s *Server) pushMids(c *gin.Context) {
	var arg struct {
		Op   int32   `form:"operation"`
		Mids []int64 `form:"mids"`
	}
	if err := c.Bind(arg); err != nil {
		writeJSON(c, nil, RequestErr)
		return
	}
	// read message
	msg, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		writeJSON(c, nil, RequestErr)
		return
	}
	if err = s.logic.PushMids(context.TODO(), arg.Op, arg.Mids, msg); err != nil {
		writeJSON(c, nil, RequestErr)
		return
	}
	writeJSON(c, nil, OK)
}

func (s *Server) pushRoom(c *gin.Context) {
	var arg struct {
		Op   int32  `form:"operation"`
		Room string `form:"room"`
	}
	if err := c.Bind(arg); err != nil {
		writeJSON(c, nil, RequestErr)
		return
	}
	// read message
	msg, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		writeJSON(c, nil, RequestErr)
		return
	}
	if err = s.logic.PushRoom(context.TODO(), arg.Op, arg.Room, msg); err != nil {
		writeJSON(c, nil, RequestErr)
		return
	}
	writeJSON(c, nil, OK)
}

func (s *Server) pushAll(c *gin.Context) {
	var arg struct {
		Op    int32  `form:"operation"`
		Speed int32  `form:"speed"`
		Tag   string `form:"tag"`
	}
	if err := c.Bind(arg); err != nil {
		writeJSON(c, nil, RequestErr)
		return
	}
	// read message
	msg, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		writeJSON(c, nil, RequestErr)
		return
	}
	if err = s.logic.PushAll(c, arg.Op, arg.Speed, arg.Tag, msg); err != nil {
		writeJSON(c, nil, RequestErr)
		return
	}
	writeJSON(c, nil, OK)
}

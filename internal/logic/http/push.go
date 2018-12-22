package http

import (
	"context"
	"io/ioutil"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (s *Server) pushKeys(c *gin.Context) {
	op, err := strconv.ParseInt(c.Query("operation"), 10, 32)
	if err != nil {
		writeJSON(c, nil, RequestErr)
		return
	}
	keys := c.QueryArray("keys")
	// read message
	msg, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		writeJSON(c, nil, RequestErr)
		return
	}
	if err = s.logic.PushKeys(context.TODO(), int32(op), keys, msg); err != nil {
		writeJSON(c, nil, RequestErr)
		return
	}
	writeJSON(c, nil, OK)
}

func (s *Server) pushMids(c *gin.Context) {
	op, err := strconv.ParseInt(c.Query("operation"), 10, 32)
	if err != nil {
		writeJSON(c, nil, RequestErr)
		return
	}
	var mids []int64
	for _, s := range c.QueryArray("mids") {
		mid, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			writeJSON(c, nil, RequestErr)
			return
		}
		mids = append(mids, mid)
	}
	if len(mids) == 0 {
		writeJSON(c, nil, RequestErr)
		return
	}
	// read message
	msg, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		writeJSON(c, nil, RequestErr)
		return
	}
	if err = s.logic.PushMids(context.TODO(), int32(op), mids, msg); err != nil {
		writeJSON(c, nil, RequestErr)
		return
	}
	writeJSON(c, nil, OK)
}

func (s *Server) pushRoom(c *gin.Context) {
	op, err := strconv.ParseInt(c.Query("operation"), 10, 32)
	if err != nil {
		writeJSON(c, nil, RequestErr)
		return
	}
	room := c.Query("room")
	if room == "" {
		writeJSON(c, nil, RequestErr)
		return
	}
	// read message
	msg, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		writeJSON(c, nil, RequestErr)
		return
	}
	if err = s.logic.PushRoom(context.TODO(), int32(op), room, msg); err != nil {
		writeJSON(c, nil, RequestErr)
		return
	}
	writeJSON(c, nil, OK)
}

func (s *Server) pushAll(c *gin.Context) {
	op, err := strconv.ParseInt(c.Query("operation"), 10, 32)
	if err != nil {
		writeJSON(c, nil, RequestErr)
		return
	}
	speed, err := strconv.ParseInt(c.Query("speed"), 10, 32)
	if err != nil {
		speed = 0
	}
	tag := c.Query("tag")
	msg, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		writeJSON(c, nil, RequestErr)
		return
	}
	if err = s.logic.PushAll(c, int32(op), int32(speed), tag, msg); err != nil {
		writeJSON(c, nil, RequestErr)
		return
	}
	writeJSON(c, nil, OK)
}

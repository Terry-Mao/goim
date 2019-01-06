package http

import (
	"github.com/gin-gonic/gin"
)

const (
	// OK ok
	OK = 0
	// RequestErr request error
	RequestErr = -400
	// ServerErr server error
	ServerErr = -500

	contextErrCode = "context/err/code"
)

type resp struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func errors(c *gin.Context, code int, msg string) {
	c.Set(contextErrCode, code)
	c.JSON(200, resp{
		Code:    code,
		Message: msg,
	})
}

func result(c *gin.Context, data interface{}, code int) {
	c.Set(contextErrCode, code)
	c.JSON(200, resp{
		Code: code,
		Data: data,
	})
}

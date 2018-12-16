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

// Ret ret.
type ret struct {
	Code int         `json:"code"`
	Data interface{} `json:"data,omitempty"`
}

func writeJSON(c *gin.Context, data interface{}, code int) (err error) {
	c.Set(contextErrCode, code)
	c.JSON(200, ret{
		Code: code,
		Data: data,
	})
	return
}

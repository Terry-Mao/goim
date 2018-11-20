package http

import (
	"encoding/json"
	"net/http"
)

const (
	// OK ok
	OK = 0
	// RequestErr request error
	RequestErr = -400
	// ServerErr server error
	ServerErr = -500
)

// Ret ret.
type ret struct {
	Code int         `json:"code"`
	Data interface{} `json:"data,omitempty"`
}

func writeJSON(w http.ResponseWriter, code int, data interface{}) (err error) {
	// write header
	header := w.Header()
	header["Content-Type"] = []string{"application/json; charset=utf-8"}
	// write body
	ret := ret{
		Code: code,
		Data: data,
	}
	b, err := json.Marshal(ret)
	if err != nil {
		return
	}
	_, err = w.Write(b)
	return
}

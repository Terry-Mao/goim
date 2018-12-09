package http

import (
	"context"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	xstrings "github.com/Terry-Mao/goim/pkg/strings"
)

func (s *Server) pushKeys(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	opStr := query.Get("operation")
	keysStr := query.Get("keys")
	// read message
	msg, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, RequestErr, nil)
		return
	}
	op, err := strconv.ParseInt(opStr, 10, 32)
	if err != nil {
		writeJSON(w, RequestErr, nil)
		return
	}
	if err = s.logic.PushKeys(context.TODO(), int32(op), strings.Split(keysStr, ","), msg); err != nil {
		writeJSON(w, RequestErr, nil)
		return
	}
	writeJSON(w, OK, nil)
}

func (s *Server) pushMids(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	opStr := query.Get("operation")
	midsStr := query.Get("mids")
	// read message
	msg, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, RequestErr, nil)
		return
	}
	op, err := strconv.ParseInt(opStr, 10, 32)
	if err != nil {
		writeJSON(w, RequestErr, nil)
		return
	}
	mids, err := xstrings.SplitInt64s(midsStr, ",")
	if err != nil {
		writeJSON(w, RequestErr, nil)
		return
	}
	if err = s.logic.PushMids(context.TODO(), int32(op), mids, msg); err != nil {
		writeJSON(w, RequestErr, nil)
		return
	}
	writeJSON(w, OK, nil)
}

func (s *Server) pushRoom(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	opStr := query.Get("op")
	room := query.Get("room")
	// read message
	msg, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, RequestErr, nil)
		return
	}
	op, err := strconv.ParseInt(opStr, 10, 32)
	if err != nil {
		writeJSON(w, RequestErr, nil)
		return
	}
	if err = s.logic.PushRoom(context.TODO(), int32(op), room, msg); err != nil {
		writeJSON(w, RequestErr, nil)
		return
	}
	writeJSON(w, OK, nil)
}

func (s *Server) pushAll(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	opStr := query.Get("operation")
	speedStr := query.Get("speed")
	tagStr := query.Get("tag")
	// read message
	msg, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, RequestErr, nil)
		return
	}
	op, err := strconv.ParseInt(opStr, 10, 32)
	if err != nil {
		writeJSON(w, RequestErr, nil)
		return
	}
	speed, err := strconv.ParseInt(speedStr, 10, 32)
	if err != nil {
		writeJSON(w, RequestErr, nil)
		return
	}
	if err = s.logic.PushAll(context.TODO(), int32(op), int32(speed), tagStr, msg); err != nil {
		writeJSON(w, RequestErr, nil)
		return
	}
	writeJSON(w, OK, nil)
}

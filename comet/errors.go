package main

import (
	"errors"
)

var (
	// server
	ErrHandshake = errors.New("handshake failed")
	ErrOperation = errors.New("request operation not valid")
	// codec
	ErrProtoPackLen   = errors.New("default server codec pack length error")
	ErrProtoHeaderLen = errors.New("default server codec header length error")
	// ring
	ErrRingEmpty = errors.New("ring buffer empty")
	ErrRingFull  = errors.New("ring buffer full")
	// timer
	ErrTimerFull   = errors.New("timer full")
	ErrTimerEmpty  = errors.New("timer empty")
	ErrTimerNoItem = errors.New("timer item not exist")
	// channel
	ErrPushArgs = errors.New("channel pushs args error")
	// crypto
	ErrInputTextSize = errors.New("input not full blocks")
)

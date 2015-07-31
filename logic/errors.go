package main

import (
	"errors"
)

var (
	ErrDecodeKey = errors.New("decode key error")
	ErrArgs      = errors.New("rpc args error")
)

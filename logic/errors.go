package main

import (
	"errors"
)

var (
	ErrRPCConfig = errors.New("rpc addrs len != networks")
	ErrDecodeKey = errors.New("decode key error")
	ErrArgs      = errors.New("rpc args error")
)

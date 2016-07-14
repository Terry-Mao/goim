package main

import (
	"errors"
)

var (
	// comet
	ErrComet     = errors.New("comet rpc is not available")
	ErrCometFull = errors.New("comet proto chan full")
	// room
	ErrRoomFull = errors.New("room proto chan full")
)

package main

import (
	"errors"
)

var (
	ErrRouter = errors.New("router rpc is not available")
	ErrComet  = errors.New("comet rpc is not available")
)

// TODO:move to common path

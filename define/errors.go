package define

import (
	"errors"
)

var (
	ErrRPCConfig = errors.New("rpc addrs len != networks")
	ErrDecodeKey = errors.New("decode key error")
	ErrArgs      = errors.New("rpc args error")
	// rpc
	ErrRouter = errors.New("router rpc is not available")
	ErrComet  = errors.New("comet rpc is not available")

	ErrNetworkAddr = errors.New("network addrs error, must network@address")
)

package xrpc

import (
	"errors"
	"goim/libs/proto"
	"net"
	"net/rpc"
	"time"

	log "github.com/thinkboy/log4go"
)

const (
	dialTimeout  = 5 * time.Second
	callTimeout  = 3 * time.Second
	pingDuration = 1 * time.Second
)

var (
	ErrRpc        = errors.New("rpc is not available")
	ErrRpcTimeout = errors.New("rpc call timeout")
)

// Rpc client options.
type ClientOptions struct {
	Proto string
	Addr  string
}

// Client is rpc client.
type Client struct {
	*rpc.Client
	options ClientOptions
	quit    chan struct{}
	err     error
}

// Dial connects to an RPC server at the specified network address.
func Dial(options ClientOptions) (c *Client) {
	c = new(Client)
	c.options = options
	c.dial()
	return
}

// Dial connects to an RPC server at the specified network address.
func (c *Client) dial() (err error) {
	var conn net.Conn
	conn, err = net.DialTimeout(c.options.Proto, c.options.Addr, dialTimeout)
	if err != nil {
		log.Error("net.Dial(%s, %s), error(%v)", c.options.Proto, c.options.Addr, err)
	} else {
		c.Client = rpc.NewClient(conn)
	}
	return
}

// Call invokes the named function, waits for it to complete, and returns its error status.
func (c *Client) Call(serviceMethod string, args interface{}, reply interface{}) (err error) {
	if c.Client == nil {
		err = ErrRpc
		return
	}
	select {
	case call := <-c.Client.Go(serviceMethod, args, reply, make(chan *rpc.Call, 1)).Done:
		err = call.Error
	case <-time.After(callTimeout):
		err = ErrRpcTimeout
	}
	return
}

// Return client error.
func (c *Client) Error() error {
	return c.err
}

// Close client connection.
func (c *Client) Close() {
	c.quit <- struct{}{}
}

// ping ping the rpc connect and reconnect when has an error.
func (c *Client) Ping(serviceMethod string) {
	var (
		arg   = proto.NoArg{}
		reply = proto.NoReply{}
		err   error
	)
	for {
		select {
		case <-c.quit:
			goto closed
			return
		default:
		}
		if c.Client != nil && c.err == nil {
			// ping
			if err = c.Call(serviceMethod, &arg, &reply); err != nil {
				c.err = err
				if err != rpc.ErrShutdown {
					c.Client.Close()
				}
				log.Error("client.Call(%s, arg, reply) error(%v)", serviceMethod, err)
			}
		} else {
			// reconnect
			if err = c.dial(); err == nil {
				// reconnect ok
				c.err = nil
				log.Info("client reconnect %s ok", c.options.Addr)
			}
		}
		time.Sleep(pingDuration)
	}
closed:
	if c.Client != nil {
		c.Client.Close()
	}
}

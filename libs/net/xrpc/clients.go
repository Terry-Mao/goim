package xrpc

import (
	"errors"
)

var (
	ErrNoClient = errors.New("rpc is not available")
)

type Clients struct {
	clients []*Client
}

// Dials connects to RPC servers at the specified network address.
func Dials(options []ClientOptions) *Clients {
	clients := new(Clients)
	for _, op := range options {
		clients.clients = append(clients.clients, Dial(op))
	}
	return clients
}

// get get a available client.
func (c *Clients) get() (*Client, error) {
	for _, cli := range c.clients {
		if cli != nil && cli.Client != nil && cli.Error() == nil {
			return cli, nil
		}
	}
	return nil, ErrNoClient
}

// has a available client.
func (c *Clients) Available() (err error) {
	_, err = c.get()
	return
}

// Call invokes the named function, waits for it to complete, and returns its error status.
// this include rpc.Client.Call method, and takes a timeout.
func (c *Clients) Call(serviceMethod string, args interface{}, reply interface{}) (err error) {
	var cli *Client
	if cli, err = c.get(); err == nil {
		err = cli.Call(serviceMethod, args, reply)
	}
	return
}

// Ping the rpc connect and reconnect when has an error.
func (c *Clients) Ping(serviceMethod string) {
	for _, cli := range c.clients {
		go cli.Ping(serviceMethod)
	}
}

package protobuf

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"net/rpc"
	"time"

	"github.com/gogo/protobuf/proto"
)

type clientCodec struct {
	commConn

	methods map[string]uint32

	reqHeader  RequestHeader
	respHeader ResponseHeader

	reqHeaderBuf bytes.Buffer
	reqBodyBuf   bytes.Buffer
}

// NewClientCodec returns a new rpc.ClientCodec using Protobuf-RPC on conn.
func NewClientCodec(conn io.ReadWriteCloser) rpc.ClientCodec {
	return &clientCodec{
		commConn: commConn{
			r: bufio.NewReader(conn),
			w: bufio.NewWriter(conn),
			c: conn,
		},
		methods: make(map[string]uint32),
	}
}

func (c *clientCodec) WriteRequest(r *rpc.Request, param interface{}) error {
	var request proto.Message
	if param != nil {
		var ok bool
		if request, ok = param.(proto.Message); !ok {
			return fmt.Errorf("protorpc.ClientCodec.WriteRequest: %T does not implement proto.Message", param)
		}
	}
	header := &c.reqHeader
	header.Seq = r.Seq
	if mid, ok := c.methods[r.ServiceMethod]; ok {
		header.MethodId = mid
	} else {
		header.Method = r.ServiceMethod
		header.MethodId = uint32(len(c.methods))
		c.methods[r.ServiceMethod] = header.MethodId
	}
	// bs, err := proto.Marshal(header)
	bs, err := marshal(&c.reqHeaderBuf, header)
	if err != nil {
		return err
	}
	if err = c.sendFrame(bs); err != nil {
		return err
	}
	// bs, err = proto.Marshal(request)
	bs, err = marshal(&c.reqBodyBuf, request)
	if err != nil {
		return err
	}
	if err = c.sendFrame(bs); err != nil {
		return err
	}
	return c.w.Flush()
}

func (c *clientCodec) ReadResponseHeader(r *rpc.Response) error {
	if err := c.recvProto(&c.respHeader); err != nil {
		return err
	}
	r.Seq = c.respHeader.Seq
	r.Error = *c.respHeader.Error
	return nil
}

func (c *clientCodec) ReadResponseBody(x interface{}) error {
	var response proto.Message
	if x != nil {
		var ok bool
		response, ok = x.(proto.Message)
		if !ok {
			return fmt.Errorf("protorpc.ClientCodec.ReadResponseBody: %T does not implement proto.Message", x)
		}
	}
	if err := c.recvProto(response); err != nil {
		return err
	}
	c.respHeader.Reset()
	return nil
}

// NewClient returns a new rpc.Client to handle requests to the
// set of services at the other end of the connection.
func NewClient(conn io.ReadWriteCloser) *rpc.Client {
	return rpc.NewClientWithCodec(NewClientCodec(conn))
}

// Dial connects to a Protobuf-RPC server at the specified network address.
func Dial(network, address string) (*rpc.Client, error) {
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	return NewClient(conn), err
}

// DialTimeout connects to a Protobuf-RPC server at the specified network address.
func DialTimeout(network, address string, timeout time.Duration) (*rpc.Client, error) {
	conn, err := net.DialTimeout(network, address, timeout)
	if err != nil {
		return nil, err
	}
	return NewClient(conn), err
}

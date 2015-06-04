package protorpc

import (
	"bufio"
	"bytes"
	"fmt"
	"io"

	"github.com/gogo/protobuf/proto"
)

type pbClientCodec struct {
	commConn

	methods map[string]uint32

	reqHeaderBuf bytes.Buffer
	reqBodyBuf   bytes.Buffer
}

// NewClientCodec returns a new ClientCodec using Protobuf-RPC on conn.
func NewClientCodec(conn io.ReadWriteCloser) ClientCodec {
	return &pbClientCodec{
		commConn: commConn{
			r: bufio.NewReader(conn),
			w: bufio.NewWriter(conn),
			c: conn,
		},
		methods: make(map[string]uint32),
	}
}

func (c *pbClientCodec) WriteRequest(r *Request, param interface{}) error {
	var request proto.Message
	if param != nil {
		var ok bool
		if request, ok = param.(proto.Message); !ok {
			return fmt.Errorf("protorpc.ClientCodec.WriteRequest: %T does not implement proto.Message", param)
		}
	}
	if mid, ok := c.methods[r.ServiceMethod]; ok {
		r.MethodId = mid
		r.ServiceMethod = ""
	} else {
		r.MethodId = uint32(len(c.methods))
		c.methods[r.ServiceMethod] = r.MethodId
	}
	// bs, err := proto.Marshal(header)
	bs, err := marshal(&c.reqHeaderBuf, r)
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

func (c *pbClientCodec) ReadResponseHeader(r *Response) error {
	return c.recvProto(r)
}

func (c *pbClientCodec) ReadResponseBody(x interface{}) error {
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
	return nil
}

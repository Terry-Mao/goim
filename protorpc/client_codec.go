package protorpc

import (
	"bufio"
	"bytes"
	"fmt"

	"github.com/gogo/protobuf/proto"
	"sync"
)

type pbClientCodec struct {
	mLock   sync.Mutex
	methods map[string]uint32

	reqHeaderBuf bytes.Buffer
	reqBodyBuf   bytes.Buffer
}

// NewPbClientCodec returns a new ClientCodec using Protobuf-RPC on conn.
func NewPbClientCodec() ClientCodec {
	return &pbClientCodec{
		methods: make(map[string]uint32),
	}
}

func (c *pbClientCodec) WriteRequest(w *bufio.Writer, r *Request, param interface{}) error {
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
		c.mLock.Lock()
		r.MethodId = uint32(len(c.methods))
		c.methods[r.ServiceMethod] = r.MethodId
		c.mLock.Unlock()
	}
	// bs, err := proto.Marshal(header)
	bs, err := marshal(&c.reqHeaderBuf, r)
	if err != nil {
		return err
	}
	if err = sendFrame(w, bs); err != nil {
		return err
	}
	// bs, err = proto.Marshal(request)
	bs, err = marshal(&c.reqBodyBuf, request)
	if err != nil {
		return err
	}
	if err = sendFrame(w, bs); err != nil {
		return err
	}
	return w.Flush()
}

func (c *pbClientCodec) ReadResponseHeader(rd *bufio.Reader, r *Response) error {
	return recvProto(rd, r)
}

func (c *pbClientCodec) ReadResponseBody(rd *bufio.Reader, x interface{}) error {
	var response proto.Message
	if x != nil {
		var ok bool
		response, ok = x.(proto.Message)
		if !ok {
			return fmt.Errorf("protorpc.ClientCodec.ReadResponseBody: %T does not implement proto.Message", x)
		}
	}
	if err := recvProto(rd, response); err != nil {
		return err
	}
	return nil
}

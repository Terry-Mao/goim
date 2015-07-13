package protorpc

import (
	"bufio"
	"bytes"
	"github.com/gogo/protobuf/proto"
	"io"
)

type pbClientCodec struct {
	rwc      io.ReadWriteCloser
	reqBuf   bytes.Buffer
	argsBuf  bytes.Buffer
	wr       *bufio.Writer
	rr       *bufio.Reader
	packSize int32
}

// NewPbClientCodec returns a new ClientCodec using Protobuf-RPC on conn.
func NewPbClientCodec(rwc io.ReadWriteCloser, rr *bufio.Reader, wr *bufio.Writer) ClientCodec {
	p := new(pbClientCodec)
	p.rwc = rwc
	p.wr = wr
	p.rr = rr
	return p
}

func (c *pbClientCodec) WriteRequest(r *Request, p proto.Message) (err error) {
	var (
		rb, pb []byte
	)
	rb, err = marshal(&c.reqBuf, r)
	if err != nil {
		return
	}
	pb, err = marshal(&c.argsBuf, p)
	if err != nil {
		return
	}
	if err = sendFrame(c.wr, rb, pb); err != nil {
		return
	}
	return c.wr.Flush()
}

func (c *pbClientCodec) ReadResponseHeader(r *Response) error {
	return recvFrame(c.rr, &c.packSize, r)
}

func (c *pbClientCodec) ReadResponseBody(b proto.Message) error {
	return recvFrame(c.rr, &c.packSize, b)
}

func (c *pbClientCodec) Close() error {
	return c.rwc.Close()
}

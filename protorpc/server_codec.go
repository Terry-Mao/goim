package protorpc

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"github.com/gogo/protobuf/proto"
	"io"
)

type pbServerCodec struct {
	rwc     io.ReadWriteCloser
	resBuf  bytes.Buffer
	repBuf  bytes.Buffer
	rr      *bufio.Reader
	wr      *bufio.Writer
	packBuf [binary.MaxVarintLen32]byte
}

// NewpbServerCodec returns a pbServerCodec that communicates with the ClientCodec
// on the other end of the given conn.
func NewPbServerCodec(rwc io.ReadWriteCloser, rr *bufio.Reader, wr *bufio.Writer) ServerCodec {
	return &pbServerCodec{
		rwc: rwc,
		rr:  rr,
		wr:  wr,
	}
}

func (c *pbServerCodec) ReadRequestHeader(r *Request) error {
	return recvFrame(c.rr, r)
}

func (c *pbServerCodec) ReadRequestBody(b proto.Message) error {
	return recvFrame(c.rr, b)
}

func (c *pbServerCodec) WriteResponse(r *Response, p proto.Message) (err error) {
	var (
		rb, pb []byte
	)
	rb, err = marshal(&c.resBuf, r)
	if err != nil {
		return
	}
	pb, err = marshal(&c.repBuf, p)
	if err != nil {
		return
	}
	if err = sendFrame(c.wr, c.packBuf[:], rb, pb); err != nil {
		return
	}
	return c.wr.Flush()
}

func (c *pbServerCodec) Close() error {
	return c.rwc.Close()
}

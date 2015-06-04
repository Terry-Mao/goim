package protorpc

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/gogo/protobuf/proto"
)

type commConn struct {
	w       *bufio.Writer
	r       *bufio.Reader
	c       io.Closer
	sizeBuf [binary.MaxVarintLen64]byte
}

func (c *commConn) Close() error {
	return c.c.Close()
}

func (c *commConn) sendFrame(data []byte) error {
	// Allocate enough space for the biggest uvarint
	size := c.sizeBuf[:]

	if data == nil || len(data) == 0 {
		n := binary.PutUvarint(size, uint64(0))
		return c.write(c.w, size[:n])
	}
	// Write the size and data
	n := binary.PutUvarint(size, uint64(len(data)))
	if err := c.write(c.w, size[:n]); err != nil {
		return err
	}
	return c.write(c.w, data)
}

func (c *commConn) write(w io.Writer, data []byte) error {
	for index := 0; index < len(data); {
		n, err := w.Write(data[index:])
		if err != nil {
			c.Close()
		}
		index += n
	}
	return nil
}

func (c *commConn) recvProto(m proto.Message) error {
	size, err := binary.ReadUvarint(c.r)
	if err != nil {
		return err
	}
	if size == 0 {
		return nil
	}
	if c.r.Buffered() >= int(size) {
		// Parse proto directly from the buffered data.
		data, err := c.r.Peek(int(size))
		if err != nil {
			return err
		}
		if err := proto.Unmarshal(data, m); err != nil {
			return err
		}
		// TODO(pmattis): This is a hack to advance the bufio pointer by
		// reading into the same slice that bufio.Reader.Peek
		// returned. In Go 1.5 we'll be able to use
		// bufio.Reader.Discard.
		_, err = io.ReadFull(c.r, data)
		return err
	}

	data := make([]byte, size)
	if _, err := io.ReadFull(c.r, data); err != nil {
		return err
	}
	return proto.Unmarshal(data, m)
}

type pbServerCodec struct {
	commConn

	methods []string

	respHeaderBuf bytes.Buffer
	respBodyBuf   bytes.Buffer
}

// NewpbServerCodec returns a pbServerCodec that communicates with the ClientCodec
// on the other end of the given conn.
func NewPbServerCodec(conn io.ReadWriteCloser) ServerCodec {
	return &pbServerCodec{
		commConn: commConn{
			r: bufio.NewReader(conn),
			w: bufio.NewWriter(conn),
			c: conn,
		},
	}
}

func (c *pbServerCodec) ReadRequestHeader(r *Request) error {
	err := c.recvProto(r)
	if err != nil {
		return err
	}
	if r.ServiceMethod == "" {
		if int(r.MethodId) >= len(c.methods) {
			return fmt.Errorf("unexpected method-id: %d >= %d", r.MethodId, len(c.methods))
		}
		r.ServiceMethod = c.methods[r.MethodId]
	} else if int(r.MethodId) > len(c.methods) {
		return fmt.Errorf("unexpected method-id: %d > %d", r.MethodId, len(c.methods))
	} else if int(r.MethodId) == len(c.methods) {
		c.methods = append(c.methods, r.ServiceMethod)
	}
	return nil
}

func (c *pbServerCodec) ReadRequestBody(x interface{}) error {
	if x == nil {
		return nil
	}
	body, ok := x.(proto.Message)
	if !ok {
		return fmt.Errorf("protorpc.pbServerCodec.ReadRequestBody: %T does not implement proto.Message", x)
	}
	err := c.recvProto(body)
	if err != nil {
		return err
	}
	return nil
}

func (c *pbServerCodec) WriteResponse(r *Response, x interface{}) error {
	var response proto.Message
	if x != nil {
		var ok bool
		if response, ok = x.(proto.Message); !ok {
			if _, ok = x.(struct{}); !ok {
				return fmt.Errorf("protorpc.pbServerCodec.WriteResponse: %T does not implement proto.Message", x)
			}
		}
	}
	// bs, err := proto.Marshal(header)
	bs, err := marshal(&c.respHeaderBuf, r)
	if err != nil {
		return err
	}
	if err = c.sendFrame(bs); err != nil {
		return err
	}
	if r.Error != "" {
		bs = nil
	} else {
		// bs, err = proto.Marshal(response)
		bs, err = marshal(&c.respBodyBuf, response)
		if err != nil {
			return err
		}
	}
	if err = c.sendFrame(bs); err != nil {
		return err
	}
	return c.w.Flush()
}

type marshalTo interface {
	Size() int
	MarshalTo([]byte) (int, error)
}

func marshal(buf *bytes.Buffer, m proto.Message) ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	if mt, ok := m.(marshalTo); ok {
		buf.Reset()
		size := mt.Size()
		buf.Grow(size)
		b := buf.Bytes()[:size]
		n, err := mt.MarshalTo(b)
		return b[:n], err
	}
	return proto.Marshal(m)
}

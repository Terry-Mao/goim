package protorpc

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/gogo/protobuf/proto"
	"sync"
)

var (
	lenPool = sync.Pool{
		New: func() interface{} {
			return make([]byte, binary.MaxVarintLen64)
		},
	}
)

type pbServerCodec struct {
	mLock   sync.Mutex
	methods []string

	bufPool sync.Pool
}

// NewpbServerCodec returns a pbServerCodec that communicates with the ClientCodec
// on the other end of the given conn.
func NewPbServerCodec() ServerCodec {
	return &pbServerCodec{
		bufPool: sync.Pool{
			New: func() interface{} {
				return bytes.NewBuffer([]byte{})
			},
		},
	}
}

func (c *pbServerCodec) ReadRequestHeader(rd *bufio.Reader, r *Request) error {
	err := recvProto(rd, r)
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
		c.mLock.Lock()
		c.methods = append(c.methods, r.ServiceMethod)
		c.mLock.Unlock()
	}
	return nil
}

func (c *pbServerCodec) ReadRequestBody(rd *bufio.Reader, x interface{}) error {
	if x == nil {
		return nil
	}
	body, ok := x.(proto.Message)
	if !ok {
		return fmt.Errorf("protorpc.pbServerCodec.ReadRequestBody: %T does not implement proto.Message", x)
	}
	err := recvProto(rd, body)
	if err != nil {
		return err
	}
	return nil
}

func (c *pbServerCodec) WriteResponse(w *bufio.Writer, cs io.Closer, r *Response, x interface{}) error {
	var response proto.Message
	if x != nil {
		var ok bool
		if response, ok = x.(proto.Message); !ok {
			if _, ok = x.(struct{}); !ok {
				return fmt.Errorf("protorpc.pbServerCodec.WriteResponse: %T does not implement proto.Message", x)
			}
		}
	}
	buf := c.bufPool.Get().(*bytes.Buffer)
	// bs, err := proto.Marshal(header)
	bs, err := marshal(buf, r)
	buf.Reset()
	if err != nil {
		c.bufPool.Put(buf)
		return err
	}
	if err = sendFrame(w, bs); err != nil {
		c.bufPool.Put(buf)
		cs.Close()
		return err
	}
	if r.Error != "" {
		bs = nil
	} else {
		// bs, err = proto.Marshal(response)
		bs, err = marshal(buf, response)
		buf.Reset()
		if err != nil {
			c.bufPool.Put(buf)
			return err
		}
	}
	if err = sendFrame(w, bs); err != nil {
		c.bufPool.Put(buf)
		cs.Close()
		return err
	}
	c.bufPool.Put(buf)
	return w.Flush()
}

func recvProto(rd *bufio.Reader, m proto.Message) error {
	size, err := binary.ReadUvarint(rd)
	if err != nil {
		return err
	}
	if size == 0 {
		return nil
	}
	if rd.Buffered() >= int(size) {
		// Parse proto directly from the buffered data.
		data, err := rd.Peek(int(size))
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
		_, err = io.ReadFull(rd, data)
		return err
	}

	data := make([]byte, size)
	if _, err := io.ReadFull(rd, data); err != nil {
		return err
	}
	return proto.Unmarshal(data, m)
}

func sendFrame(w *bufio.Writer, data []byte) error {
	// Allocate enough space for the biggest uvarint
	size := lenPool.Get().([]byte)

	if data == nil || len(data) == 0 {
		n := binary.PutUvarint(size, uint64(0))
		err := write(w, size[:n])
		lenPool.Put(size)
		return err
	}
	// Write the size and data
	n := binary.PutUvarint(size, uint64(len(data)))
	if err := write(w, size[:n]); err != nil {
		lenPool.Put(size)
		return err
	}
	lenPool.Put(size)
	return write(w, data)
}

func write(w io.Writer, data []byte) error {
	for index := 0; index < len(data); {
		n, err := w.Write(data[index:])
		if err != nil {
			return err
		}
		index += n
	}
	return nil
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

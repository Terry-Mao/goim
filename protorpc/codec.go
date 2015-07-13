package protorpc

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"github.com/gogo/protobuf/proto"
	"io"
)

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

func sendFrame(wr *bufio.Writer, b []byte, ds ...[]byte) (err error) {
	var n int
	// size & data
	for _, d := range ds {
		n = binary.PutVarint(b, int64(len(d)))
		if _, err = wr.Write(b[:n]); err != nil {
			return
		}
		if _, err = wr.Write(d); err != nil {
			return
		}
	}
	return
}

func recvFrame(rd *bufio.Reader, m proto.Message) (err error) {
	var (
		d     []byte
		vsize int64
		size  int
	)
	if vsize, err = binary.ReadVarint(rd); err != nil || vsize == 0 {
		return
	} else {
		size = int(vsize)
	}
	if rd.Buffered() >= size {
		// Parse proto directly from the buffered data.
		if d, err = rd.Peek(size); err != nil {
			return
		}
		// simply discard
		if m != nil {
			if err = proto.Unmarshal(d, m); err != nil {
				return
			}
		}
		// TODO(pmattis): This is a hack to advance the bufio pointer by
		// reading into the same slice that bufio.Reader.Peek
		// returned. In Go 1.5 we'll be able to use
		// bufio.Reader.Discard.
		_, err = io.ReadFull(rd, d)
		return
	}
	d = make([]byte, size)
	if _, err = io.ReadFull(rd, d); err != nil {
		return
	}
	// simply discard
	if m != nil {
		return proto.Unmarshal(d, m)
	}
	return
}

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

func sendFrame(wr *bufio.Writer, ds ...[]byte) (err error) {
	// size & data
	for _, d := range ds {
		if err = binary.Write(wr, binary.BigEndian, int32(len(d))); err != nil {
			return
		} else {
			if err = binary.Write(wr, binary.BigEndian, d); err != nil {
				return
			}
		}
	}
	return
}

func recvFrame(rd *bufio.Reader, m proto.Message) (err error) {
	var (
		size int32
		d    []byte
	)
	if err = binary.Read(rd, binary.BigEndian, &size); err != nil {
		return
	} else if size == 0 {
		return
	}
	if rd.Buffered() >= int(size) {
		// Parse proto directly from the buffered data.
		if d, err = rd.Peek(int(size)); err != nil {
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

package proto

import (
	"encoding/json"
	"errors"
	"fmt"
	"goim/libs/bufio"
	"goim/libs/bytes"
	"goim/libs/define"
	"goim/libs/encoding/binary"

	"github.com/gorilla/websocket"
)

// for tcp
const (
	MaxBodySize = int32(1 << 10)
)

const (
	// size
	PackSize      = 4
	HeaderSize    = 2
	VerSize       = 2
	OperationSize = 4
	SeqIdSize     = 4
	RawHeaderSize = PackSize + HeaderSize + VerSize + OperationSize + SeqIdSize
	MaxPackSize   = MaxBodySize + int32(RawHeaderSize)
	// offset
	PackOffset      = 0
	HeaderOffset    = PackOffset + PackSize
	VerOffset       = HeaderOffset + HeaderSize
	OperationOffset = VerOffset + VerSize
	SeqIdOffset     = OperationOffset + OperationSize
)

var (
	emptyProto    = Proto{}
	emptyJSONBody = []byte("{}")

	ErrProtoPackLen   = errors.New("default server codec pack length error")
	ErrProtoHeaderLen = errors.New("default server codec header length error")
)

var (
	ProtoReady  = &Proto{Operation: define.OP_PROTO_READY}
	ProtoFinish = &Proto{Operation: define.OP_PROTO_FINISH}
)

// Proto is a request&response written before every goim connect.  It is used internally
// but documented here as an aid to debugging, such as when analyzing
// network traffic.
// tcp:
// binary codec
// websocket & http:
// raw codec, with http header stored ver, operation, seqid
type Proto struct {
	Ver       int16           `json:"ver"`  // protocol version
	Operation int32           `json:"op"`   // operation for request
	SeqId     int32           `json:"seq"`  // sequence number chosen by client
	Body      json.RawMessage `json:"body"` // binary body bytes(json.RawMessage is []byte)
}

func (p *Proto) Reset() {
	*p = emptyProto
}

func (p *Proto) String() string {
	return fmt.Sprintf("\n-------- proto --------\nver: %d\nop: %d\nseq: %d\nbody: %v\n-----------------------", p.Ver, p.Operation, p.SeqId, p.Body)
}

func (p *Proto) WriteTo(b *bytes.Writer) {
	var (
		packLen = RawHeaderSize + int32(len(p.Body))
		buf     = b.Peek(RawHeaderSize)
	)
	binary.BigEndian.PutInt32(buf[PackOffset:], packLen)
	binary.BigEndian.PutInt16(buf[HeaderOffset:], int16(RawHeaderSize))
	binary.BigEndian.PutInt16(buf[VerOffset:], p.Ver)
	binary.BigEndian.PutInt32(buf[OperationOffset:], p.Operation)
	binary.BigEndian.PutInt32(buf[SeqIdOffset:], p.SeqId)
	if p.Body != nil {
		b.Write(p.Body)
	}
}

func (p *Proto) ReadTCP(rr *bufio.Reader) (err error) {
	var (
		bodyLen   int
		headerLen int16
		packLen   int32
		buf       []byte
	)
	if buf, err = rr.Pop(RawHeaderSize); err != nil {
		return
	}
	packLen = binary.BigEndian.Int32(buf[PackOffset:HeaderOffset])
	headerLen = binary.BigEndian.Int16(buf[HeaderOffset:VerOffset])
	p.Ver = binary.BigEndian.Int16(buf[VerOffset:OperationOffset])
	p.Operation = binary.BigEndian.Int32(buf[OperationOffset:SeqIdOffset])
	p.SeqId = binary.BigEndian.Int32(buf[SeqIdOffset:])
	if packLen > MaxPackSize {
		return ErrProtoPackLen
	}
	if headerLen != RawHeaderSize {
		return ErrProtoHeaderLen
	}
	if bodyLen = int(packLen - int32(headerLen)); bodyLen > 0 {
		p.Body, err = rr.Pop(bodyLen)
	} else {
		p.Body = nil
	}
	return
}

func (p *Proto) WriteTCP(wr *bufio.Writer) (err error) {
	var (
		buf     []byte
		packLen int32
	)
	if p.Operation == define.OP_RAW {
		// write without buffer, job concact proto into raw buffer
		_, err = wr.WriteRaw(p.Body)
		return
	}
	packLen = RawHeaderSize + int32(len(p.Body))
	if buf, err = wr.Peek(RawHeaderSize); err != nil {
		return
	}
	binary.BigEndian.PutInt32(buf[PackOffset:], packLen)
	binary.BigEndian.PutInt16(buf[HeaderOffset:], int16(RawHeaderSize))
	binary.BigEndian.PutInt16(buf[VerOffset:], p.Ver)
	binary.BigEndian.PutInt32(buf[OperationOffset:], p.Operation)
	binary.BigEndian.PutInt32(buf[SeqIdOffset:], p.SeqId)
	if p.Body != nil {
		_, err = wr.Write(p.Body)
	}
	return
}

func (p *Proto) ReadWebsocket(wr *websocket.Conn) (err error) {
	err = wr.ReadJSON(p)
	return
}

func (p *Proto) WriteBodyTo(b *bytes.Writer) (err error) {
	var (
		ph  Proto
		js  []json.RawMessage
		j   json.RawMessage
		jb  []byte
		bts []byte
	)
	offset := int32(PackOffset)
	buf := p.Body[:]
	for {
		if (len(buf[offset:])) < RawHeaderSize {
			// should not be here
			break
		}
		packLen := binary.BigEndian.Int32(buf[offset : offset+HeaderOffset])
		packBuf := buf[offset : offset+packLen]
		// packet
		ph.Ver = binary.BigEndian.Int16(packBuf[VerOffset:OperationOffset])
		ph.Operation = binary.BigEndian.Int32(packBuf[OperationOffset:SeqIdOffset])
		ph.SeqId = binary.BigEndian.Int32(packBuf[SeqIdOffset:RawHeaderSize])
		ph.Body = packBuf[RawHeaderSize:]
		if jb, err = json.Marshal(&ph); err != nil {
			return
		}
		j = json.RawMessage(jb)
		js = append(js, j)
		offset += packLen
	}
	if bts, err = json.Marshal(js); err != nil {
		return
	}
	b.Write(bts)
	return
}

func (p *Proto) WriteWebsocket(wr *websocket.Conn) (err error) {
	if p.Body == nil {
		p.Body = emptyJSONBody
	}
	// [{"ver":1,"op":8,"seq":1,"body":{}}, {"ver":1,"op":3,"seq":2,"body":{}}]
	if p.Operation == define.OP_RAW {
		// batch mod
		var b = bytes.NewWriterSize(len(p.Body) + 40*RawHeaderSize)
		if err = p.WriteBodyTo(b); err != nil {
			return
		}
		err = wr.WriteMessage(websocket.TextMessage, b.Buffer())
		return
	}
	err = wr.WriteJSON([]*Proto{p})
	return
}

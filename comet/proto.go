package main

import (
	"encoding/json"
	"fmt"
)

// for tcp
const (
	rawHeaderLen  = int16(16)
	maxBodyLen    = int32(1 << 10)
	maxPackLen    = maxBodyLen + int32(rawHeaderLen)
	packLenSize   = 4
	headerLenSize = 2
	maxPackIntBuf = 4
)

const (
	VerSize       = 2
	OperationSize = 4
	SeqIdSize     = 4
)

var (
	emptyProto = Proto{}
)

// Proto is a request&response written before every goim connect.  It is used internally
// but documented here as an aid to debugging, such as when analyzing
// network traffic.
// tcp:
// binary codec
// websocket & http:
// raw codec, with http header stored ver, operation, seqid
type Proto struct {
	Ver       int16            `json:"ver"`  // protocol version
	Operation int32            `json:"op"`   // operation for request
	SeqId     int32            `json:"seq"`  // sequence number chosen by client
	Body      json.RawMessage  `json:"body"` // binary body bytes(json.RawMessage is []byte)
	Buf       [maxBodyLen]byte `json:"-"`    // for upstream buf
}

func (p *Proto) Reset() {
	*p = emptyProto
}

func (p *Proto) String() string {
	return fmt.Sprintf("\n-------- proto --------\nver: %d\nop: %d\nseq: %d\nbody: %s\n-----------------------", p.Ver, p.Operation, p.SeqId, string(p.Body))
}

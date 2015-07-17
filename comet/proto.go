package main

import (
	log "code.google.com/p/log4go"
	"encoding/json"
)

const (
	VerSize       = 2
	OperationSize = 4
	SeqIdSize     = 4
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

func (p *Proto) Print() {
	log.Debug("\n-------- proto --------\nver: %d\nop: %d\nseq: %d\nbody: %s\n", p.Ver, p.Operation, p.SeqId, string(p.Body))
}

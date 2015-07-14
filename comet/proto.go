package main

const (
	PackLenSize   = 4
	HeaderLenSize = 2
	VerSize       = 2
	OperationSize = 4
	SeqIdSize     = 4
)

// Proto is a request&response written before every goim connect.  It is used internally
// but documented here as an aid to debugging, such as when analyzing
// network traffic.
type Proto struct {
	PackLen   int32  // package length
	HeaderLen int16  // header length
	Ver       int16  // protocol version
	Operation int32  // operation for request
	SeqId     int32  // sequence number chosen by client
	Body      []byte // body bytes
}

package main

import (
	log "code.google.com/p/log4go"
	"io"
)

const (
	maxPackLen    = 1 << 10
	maxPackIntBuf = 4
	rawHeaderLen  = int16(16)
)

type Flusher interface {
	Flush() error
}

type ServerCodec interface {
	ReadRequestHeader(io.Reader, *Proto) error
	ReadRequestBody(io.Reader, *Proto) error
	// WriteResponse must be safe for concurrent use by multiple goroutines.
	WriteResponse(io.Writer, Flusher, *Proto) error
}

type BinaryServerCodec struct {
	packIntBuf [maxPackIntBuf]byte
}

func (c *BinaryServerCodec) ReadRequestHeader(rd io.Reader, proto *Proto) (err error) {
	if err = ReadAll(rd, c.packIntBuf[:PackLenSize]); err != nil {
		return
	}
	proto.PackLen = BigEndian.Int32(c.packIntBuf[:PackLenSize])
	log.Debug("packLen: %d", proto.PackLen)
	if proto.PackLen > maxPackLen {
		return ErrProtoPackLen
	}
	if err = ReadAll(rd, c.packIntBuf[:HeaderLenSize]); err != nil {
		return
	}
	proto.HeaderLen = BigEndian.Int16(c.packIntBuf[:HeaderLenSize])
	log.Debug("headerLen: %d", proto.HeaderLen)
	if proto.HeaderLen != rawHeaderLen {
		return ErrProtoHeaderLen
	}
	if err = ReadAll(rd, c.packIntBuf[:VerSize]); err != nil {
		return
	}
	proto.Ver = BigEndian.Int16(c.packIntBuf[:VerSize])
	log.Debug("protoVer: %d", proto.Ver)
	if err = ReadAll(rd, c.packIntBuf[:OperationSize]); err != nil {
		return
	}
	proto.Operation = BigEndian.Int32(c.packIntBuf[:OperationSize])
	log.Debug("operation: %d", proto.Operation)
	if err = ReadAll(rd, c.packIntBuf[:SeqIdSize]); err != nil {
		return
	}
	proto.SeqId = BigEndian.Int32(c.packIntBuf[:SeqIdSize])
	log.Debug("seqId: %d", proto.SeqId)
	return
}

func (c *BinaryServerCodec) ReadRequestBody(rd io.Reader, proto *Proto) (err error) {
	bodyLen := int(proto.PackLen - int32(proto.HeaderLen))
	log.Debug("read body len: %d", bodyLen)
	if bodyLen > 0 {
		proto.Body = make([]byte, bodyLen)
		if err = ReadAll(rd, proto.Body); err != nil {
			log.Error("body: ReadAll() error(%v)", err)
			return
		}
	} else {
		proto.Body = nil
	}
	return
}

func (c *BinaryServerCodec) WriteResponse(wr io.Writer, fr Flusher, proto *Proto) (err error) {
	log.Debug("write proto: %v", proto)
	BigEndian.PutInt32(c.packIntBuf[:PackLenSize], int32(rawHeaderLen)+int32(len(proto.Body)))
	if _, err = wr.Write(c.packIntBuf[:PackLenSize]); err != nil {
		log.Error("packLen: wr.Write() error(%v)", err)
		return
	}
	BigEndian.PutInt16(c.packIntBuf[:HeaderLenSize], rawHeaderLen)
	if _, err = wr.Write(c.packIntBuf[:HeaderLenSize]); err != nil {
		log.Error("headerLen: wr.Write() error(%v)", err)
		return
	}
	BigEndian.PutInt16(c.packIntBuf[:VerSize], proto.Ver)
	if _, err = wr.Write(c.packIntBuf[:VerSize]); err != nil {
		log.Error("protoVer: wr.Write() error(%v)", err)
		return
	}
	BigEndian.PutInt32(c.packIntBuf[:OperationSize], proto.Operation)
	if _, err = wr.Write(c.packIntBuf[:OperationSize]); err != nil {
		log.Error("operation: wr.Write() error(%v)", err)
		return
	}
	BigEndian.PutInt32(c.packIntBuf[:SeqIdSize], proto.SeqId)
	if _, err = wr.Write(c.packIntBuf[:SeqIdSize]); err != nil {
		log.Error("seqId: wr.Write() error(%v)", err)
		return
	}
	if proto.Body != nil {
		if _, err = wr.Write(proto.Body); err != nil {
			log.Error("body: wr.Write() error(%v)", err)
			return
		}
	}
	if fr != nil {
		return fr.Flush()
	}
	return
}

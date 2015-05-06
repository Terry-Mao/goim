package main

import (
	"bufio"
	log "code.google.com/p/log4go"
	"encoding/binary"
	"errors"
)

const (
	maxPackLen   = 2 << 10
	rawHeaderLen = int16(16)
)

var (
	ErrProtoPackLen   = errors.New("default server codec pack length error")
	ErrProtoHeaderLen = errors.New("default server codec header length error")
)

type DefaultServerCodec struct {
}

func (c *DefaultServerCodec) ReadRequestHeader(rd *bufio.Reader, proto *Proto) (err error) {
	if err = binary.Read(rd, binary.BigEndian, &proto.PackLen); err != nil {
		log.Error("packLen: binary.Read() error(%v)", err)
		return
	}
	log.Debug("packLen: %d", proto.PackLen)
	if proto.PackLen > maxPackLen {
		return ErrProtoPackLen
	}
	if err = binary.Read(rd, binary.BigEndian, &proto.HeaderLen); err != nil {
		log.Error("headerLen: binary.Read() error(%v)", err)
		return
	}
	log.Debug("headerLen: %d", proto.HeaderLen)
	if proto.HeaderLen != rawHeaderLen {
		return ErrProtoHeaderLen
	}
	if err = binary.Read(rd, binary.BigEndian, &proto.Ver); err != nil {
		log.Error("protoVer: binary.Read() error(%v)", err)
		return
	}
	// TODO check ver
	log.Debug("protoVer: %d", proto.Ver)
	if err = binary.Read(rd, binary.BigEndian, &proto.Operation); err != nil {
		log.Error("Operation: binary.Read() error(%v)", err)
		return
	}
	log.Debug("operation: %d", proto.Operation)
	if err = binary.Read(rd, binary.BigEndian, &proto.SeqId); err != nil {
		log.Error("seqId: binary.Read() error(%v)", err)
		return
	}
	log.Debug("seqId: %d", proto.SeqId)
	return
}

func (c *DefaultServerCodec) ReadRequestBody(rd *bufio.Reader, proto *Proto) (err error) {
	var (
		n       = int(0)
		t       = int(0)
		bodyLen = int(proto.PackLen - int32(proto.HeaderLen))
	)
	log.Debug("read body len: %d", bodyLen)
	if bodyLen > 0 {
		proto.Body = make([]byte, bodyLen)
		// no deadline, because readheader always incoming calls readbody
		for {
			if t, err = rd.Read(proto.Body[n:]); err != nil {
				log.Error("body: buf.Read() error(%v)", err)
				return
			}
			if n += t; n == bodyLen {
				log.Debug("body: rd.Read() fill ok")
				break
			} else if n < bodyLen {
				log.Debug("body: rd.Read() need %d bytes", bodyLen-n)
			} else {
				log.Error("body: readbytes %d > %d", n, bodyLen)
			}
		}
	} else {
		proto.Body = nil
	}
	return
}

func (c *DefaultServerCodec) WriteResponse(wr *bufio.Writer, proto *Proto) (err error) {
	log.Debug("write proto: %v", proto)
	// packlen =header(16) + body
	if err = binary.Write(wr, binary.BigEndian, uint32(rawHeaderLen)+uint32(len(proto.Body))); err != nil {
		log.Error("packLen: binary.Write() error(%v)", err)
		return
	}
	if err = binary.Write(wr, binary.BigEndian, rawHeaderLen); err != nil {
		log.Error("headerLen: binary.Write() error(%v)", err)
		return
	}
	if err = binary.Write(wr, binary.BigEndian, proto.Ver); err != nil {
		log.Error("protoVer: binary.Write() error(%v)", err)
		return
	}
	if err = binary.Write(wr, binary.BigEndian, proto.Operation); err != nil {
		log.Error("operation: binary.Write() error(%v)", err)
		return
	}
	if err = binary.Write(wr, binary.BigEndian, proto.SeqId); err != nil {
		log.Error("seqId: binary.Write() error(%v)", err)
		return
	}
	if proto.Body != nil {
		if err = binary.Write(wr, binary.BigEndian, proto.Body); err != nil {
			log.Error("body: binary.Write() error(%v)", err)
			return
		}
	}
	return wr.Flush()
}

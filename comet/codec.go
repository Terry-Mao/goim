package main

import (
	"bufio"
	log "code.google.com/p/log4go"
	"encoding/binary"
	"net"
	"time"
)

const (
	packBytes    = 4
	headerBytes  = 2
	rawPackLen   = uint32(16)
	rawHeaderLen = uint16(12)
)

type IMServerCodec struct {
	conn      net.Conn
	rdBuf     *bufio.Reader
	wrBuf     *bufio.Writer
	packLen   uint32
	headerLen uint16
}

func (c *IMServerCodec) ReadRequestHeader(proto *Proto) (err error) {
	if err = c.conn.SetReadDeadline(time.Now().Add(Conf.ReadTimeout)); err != nil {
		log.Error("conn.SetReadDeadline() error(%v)", err)
		return
	}
	if err = binary.Read(c.rdBuf, binary.BigEndian, &c.packLen); err != nil {
		log.Error("packLen: binary.Read() error(%v)", err)
		return
	}
	log.Debug("packLen: %d", c.packLen)
	// TODO packlen check
	if err = binary.Read(c.rdBuf, binary.BigEndian, &c.headerLen); err != nil {
		log.Error("headerLen: binary.Read() error(%v)", err)
		return
	}
	log.Debug("headerLen: %d", c.headerLen)
	// TODO headerlen check
	if err = binary.Read(c.rdBuf, binary.BigEndian, &proto.Ver); err != nil {
		log.Error("protoVer: binary.Read() error(%v)", err)
		return
	}
	log.Debug("protoVer: %d", proto.Ver)
	if err = binary.Read(c.rdBuf, binary.BigEndian, &proto.Operation); err != nil {
		log.Error("Operation: binary.Read() error(%v)", err)
		return
	}
	log.Debug("operation: %d", proto.Operation)
	if err = binary.Read(c.rdBuf, binary.BigEndian, &proto.SeqId); err != nil {
		log.Error("seqId: binary.Read() error(%v)", err)
		return
	}
	log.Debug("seqId: %d", proto.SeqId)
	return
}

func (c *IMServerCodec) ReadRequestBody() (body []byte, err error) {
	var (
		n       = int(0)
		t       = int(0)
		bodyLen = int(c.packLen - uint32(c.headerLen) - packBytes)
	)
	log.Debug("read body len: %d", bodyLen)
	if bodyLen > 0 {
		body = make([]byte, bodyLen)
		// no deadline, because readheader always incoming calls readbody
		for {
			if t, err = c.rdBuf.Read(body[n:]); err != nil {
				log.Error("body: buf.Read() error(%v)", err)
				return
			}
			if n += t; n == bodyLen {
				log.Debug("body: c.Read() fill ok")
				break
			} else if n < bodyLen {
				log.Debug("body: c.Read() need %d bytes", bodyLen-n)
			} else {
				log.Error("body: readbytes %d > %d", n, bodyLen)
			}
		}
	}
	return
}

func (c *IMServerCodec) WriteResponse(proto *Proto) (err error) {
	if err = c.conn.SetWriteDeadline(time.Now().Add(Conf.WriteTimeout)); err != nil {
		log.Error("conn.SetWriteDeadline() error(%v)", err)
		return
	}
	log.Debug("write proto: %v", proto)
	if err = binary.Write(c.wrBuf, binary.BigEndian, rawPackLen+uint32(len(proto.Body))); err != nil {
		log.Error("packLen: binary.Write() error(%v)", err)
		return
	}
	if err = binary.Write(c.wrBuf, binary.BigEndian, rawHeaderLen); err != nil {
		log.Error("headerLen: binary.Write() error(%v)", err)
		return
	}
	if err = binary.Write(c.wrBuf, binary.BigEndian, proto.Ver); err != nil {
		log.Error("protoVer: binary.Write() error(%v)", err)
		return
	}
	if err = binary.Write(c.wrBuf, binary.BigEndian, proto.Operation); err != nil {
		log.Error("operation: binary.Write() error(%v)", err)
		return
	}
	if err = binary.Write(c.wrBuf, binary.BigEndian, proto.SeqId); err != nil {
		log.Error("seqId: binary.Write() error(%v)", err)
		return
	}
	if proto.Body != nil {
		if err = binary.Write(c.wrBuf, binary.BigEndian, proto.Body); err != nil {
			log.Error("body: binary.Write() error(%v)", err)
			return
		}
	}
	return c.wrBuf.Flush()
}

func (c *IMServerCodec) Close() error {
	return c.conn.Close()
}

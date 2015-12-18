package main

import (
	log "code.google.com/p/log4go"
	"github.com/Terry-Mao/goim/libs/bufio"
	"github.com/Terry-Mao/goim/libs/bytes"
	"github.com/Terry-Mao/goim/libs/define"
	"github.com/Terry-Mao/goim/libs/encoding/binary"
	itime "github.com/Terry-Mao/goim/libs/time"
	"net"
	"time"
)

// InitTCP listen all tcp.bind and start accept connections.
func InitTCP(addrs []string, accept int) (err error) {
	var (
		bind     string
		listener *net.TCPListener
		addr     *net.TCPAddr
	)
	for _, bind = range addrs {
		if addr, err = net.ResolveTCPAddr("tcp4", bind); err != nil {
			log.Error("net.ResolveTCPAddr(\"tcp4\", \"%s\") error(%v)", bind, err)
			return
		}
		if listener, err = net.ListenTCP("tcp4", addr); err != nil {
			log.Error("net.ListenTCP(\"tcp4\", \"%s\") error(%v)", bind, err)
			return
		}
		if Debug {
			log.Debug("start tcp listen: \"%s\"", bind)
		}
		// split N core accept
		for i := 0; i < accept; i++ {
			go acceptTCP(DefaultServer, listener)
		}
	}
	return
}

// Accept accepts connections on the listener and serves requests
// for each incoming connection.  Accept blocks; the caller typically
// invokes it in a go statement.
func acceptTCP(server *Server, lis *net.TCPListener) {
	var (
		conn *net.TCPConn
		err  error
		r    int
	)
	for {
		if conn, err = lis.AcceptTCP(); err != nil {
			// if listener close then return
			log.Error("listener.Accept(\"%s\") error(%v)", lis.Addr().String(), err)
			return
		}
		if err = conn.SetKeepAlive(server.Options.TCPKeepalive); err != nil {
			log.Error("conn.SetKeepAlive() error(%v)", err)
			return
		}
		if err = conn.SetReadBuffer(server.Options.TCPRcvbuf); err != nil {
			log.Error("conn.SetReadBuffer() error(%v)", err)
			return
		}
		if err = conn.SetWriteBuffer(server.Options.TCPSndbuf); err != nil {
			log.Error("conn.SetWriteBuffer() error(%v)", err)
			return
		}
		go serveTCP(server, conn, r)
		if r++; r == maxInt {
			r = 0
		}
	}
}

func serveTCP(server *Server, conn *net.TCPConn, r int) {
	var (
		// timer
		tr = server.round.Timer(r)
		rp = server.round.Reader(r)
		wp = server.round.Writer(r)
		// ip addr
		lAddr = conn.LocalAddr().String()
		rAddr = conn.RemoteAddr().String()
	)
	if Debug {
		log.Debug("start tcp serve \"%s\" with \"%s\"", lAddr, rAddr)
	}
	server.serveTCP(conn, rp, wp, tr)
}

func (server *Server) serveTCP(conn *net.TCPConn, rp, wp *bytes.Pool, tr *itime.Timer) {
	var (
		b       *Bucket
		key     string
		hb      time.Duration // heartbeat
		bodyLen int
		err     error
		trd     *itime.TimerData
		rb      = rp.Get()
		wb      = wp.Get()
		ch      = NewChannel(server.Options.Proto, define.NoRoom)
		rr      = &ch.Reader
		wr      = &ch.Writer
		p       = &ch.CliProto
	)
	ch.Reader.ResetBuffer(conn, rb.Bytes())
	ch.Writer.ResetBuffer(conn, wb.Bytes())
	// handshake
	trd = tr.Add(server.Options.HandshakeTimeout, func() {
		conn.Close()
	})
	if key, ch.RoomId, hb, err = server.authTCP(rr, wr, p); err != nil {
		conn.Close()
		rp.Put(rb)
		wp.Put(wb)
		tr.Del(trd)
		log.Error("key: %s handshake failed error(%v)", key, err)
		return
	}
	trd.Key = key
	tr.Set(trd, hb)
	b = server.Bucket(key)
	b.Put(key, ch, tr)
	// hanshake ok start dispatch goroutine
	go server.dispatchTCP(key, conn, wr, wp, wb, ch)
	for {
		if bodyLen, err = server.readTCPRequest(rr, p); err != nil {
			break
		}
		if p.Operation == define.OP_HEARTBEAT {
			tr.Set(trd, hb)
			p.Body = nil
			p.Operation = define.OP_HEARTBEAT_REPLY
			if Debug {
				log.Debug("key: %s receive heartbeat", key)
			}
		} else {
			if err = server.operator.Operate(p); err != nil {
				break
			}
		}
		if err = ch.Reply(p); err != nil {
			break
		}
		if _, err = rr.Discard(bodyLen); err != nil {
			break
		}
	}
	log.Error("key: %s server tcp failed error(%v)", key, err)
	conn.Close()
	ch.Close()
	rp.Put(rb)
	b.Del(key)
	tr.Del(trd)
	if err = server.operator.Disconnect(key, ch.RoomId); err != nil {
		log.Error("key: %s operator do disconnect error(%v)", key, err)
	}
	if Debug {
		log.Debug("key: %s server tcp goroutine exit", key)
	}
	return
}

// dispatch accepts connections on the listener and serves requests
// for each incoming connection.  dispatch blocks; the caller typically
// invokes it in a go statement.
func (server *Server) dispatchTCP(key string, conn *net.TCPConn, wr *bufio.Writer, wp *bytes.Pool, wb *bytes.Buffer, ch *Channel) {
	var (
		p   *Proto
		err error
	)
	if Debug {
		log.Debug("key: %s start dispatch tcp goroutine", key)
	}
	for {
		if !ch.Ready() {
			if Debug {
				log.Debug("key: %s wakeup exit dispatch goroutine", key)
			}
			break
		}
		// fetch message from svrbox(server send)
		for {
			if p, err = ch.SvrProto.Get(); err != nil {
				// must be empty error
				err = nil
				break
			}
			// just forward the message
			if err = server.writeTCPResponse(wr, p); err != nil {
				break
			}
			ch.SvrProto.GetAdv()
		}
		if err != nil {
			break
		}
		// only hungry flush response
		if err = wr.Flush(); err != nil {
			break
		}
	}
	log.Error("key: %s dispatch tcp error(%v)", key, err)
	conn.Close()
	wp.Put(wb)
	if Debug {
		log.Debug("key: %s dispatch goroutine exit", key)
	}
	return
}

// auth for goim handshake with client, use rsa & aes.
func (server *Server) authTCP(rr *bufio.Reader, wr *bufio.Writer, p *Proto) (key string, rid int32, heartbeat time.Duration, err error) {
	var (
		bodyLen int
	)
	if bodyLen, err = server.readTCPRequest(rr, p); err != nil {
		return
	}
	if p.Operation != define.OP_AUTH {
		log.Warn("auth operation not valid: %d", p.Operation)
		err = ErrOperation
		return
	}
	if key, rid, heartbeat, err = server.operator.Connect(p); err != nil {
		return
	}
	p.Body = nil
	p.Operation = define.OP_AUTH_REPLY
	if err = server.writeTCPResponse(wr, p); err != nil {
		return
	}
	if err = wr.Flush(); err != nil {
		return
	}
	_, err = rr.Discard(bodyLen)
	return
}

// readRequest
func (server *Server) readTCPRequest(rr *bufio.Reader, p *Proto) (bodyLen int, err error) {
	var (
		packLen int32
		buf     []byte
	)
	if buf, err = rr.Peek(RawHeaderSize); err != nil {
		return
	}
	packLen = binary.BigEndian.Int32(buf[PackOffset:HeaderOffset])
	p.HeaderLen = binary.BigEndian.Int16(buf[HeaderOffset:VerOffset])
	p.Ver = binary.BigEndian.Int16(buf[VerOffset:OperationOffset])
	p.Operation = binary.BigEndian.Int32(buf[OperationOffset:SeqIdOffset])
	p.SeqId = binary.BigEndian.Int32(buf[SeqIdOffset:EndOffset])
	if _, err = rr.Discard(RawHeaderSize); err != nil {
		return
	}
	if packLen > MaxPackSize {
		return 0, ErrProtoPackLen
	}
	if p.HeaderLen != RawHeaderSize {
		return 0, ErrProtoHeaderLen
	}
	if bodyLen = int(packLen - int32(p.HeaderLen)); bodyLen > 0 {
		p.Body, err = rr.Peek(bodyLen)
	} else {
		p.Body = nil
	}
	if Debug {
		log.Debug("read proto: %v", p)
	}
	return
}

// sendResponse send resp to client, sendResponse must be goroutine safe.
func (server *Server) writeTCPResponse(wr *bufio.Writer, p *Proto) (err error) {
	var (
		buf     []byte
		packLen int32
	)
	if buf, err = wr.Peek(RawHeaderSize); err != nil {
		return
	}
	packLen = RawHeaderSize + int32(len(p.Body))
	p.HeaderLen = RawHeaderSize
	binary.BigEndian.PutInt32(buf[PackOffset:], packLen)
	binary.BigEndian.PutInt16(buf[HeaderOffset:], p.HeaderLen)
	binary.BigEndian.PutInt16(buf[VerOffset:], p.Ver)
	binary.BigEndian.PutInt32(buf[OperationOffset:], p.Operation)
	binary.BigEndian.PutInt32(buf[SeqIdOffset:], p.SeqId)
	// TODO writev
	if p.Body != nil {
		_, err = wr.Write(p.Body)
	}
	if Debug {
		log.Debug("write proto: %v", p)
	}
	return
}

package main

import (
	"bufio"
	log "code.google.com/p/log4go"
	"github.com/Terry-Mao/goim/libs/define"
	"github.com/Terry-Mao/goim/libs/encoding/binary"
	"net"
	"sync"
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
		// bufpool
		rrp = server.round.Reader(r) // reader
		wrp = server.round.Writer(r) // writer
		// timer
		tr = server.round.Timer(r)
		// buf
		rr = NewBufioReaderSize(rrp, conn, server.Options.TCPReadBufSize)  // reader buf
		wr = NewBufioWriterSize(wrp, conn, server.Options.TCPWriteBufSize) // writer buf
		// ip addr
		lAddr = conn.LocalAddr().String()
		rAddr = conn.RemoteAddr().String()
	)
	if Debug {
		log.Debug("start tcp serve \"%s\" with \"%s\"", lAddr, rAddr)
	}
	server.serveTCP(conn, rrp, wrp, rr, wr, tr)
}

func (server *Server) serveTCP(conn *net.TCPConn, rrp, wrp *sync.Pool, rr *bufio.Reader, wr *bufio.Writer, tr *Timer) {
	var (
		b       *Bucket
		key     string
		hb      time.Duration // heartbeat
		bodyLen int
		err     error
		trd     *TimerData
		ch      = NewChannel(server.Options.SvrProto, define.NoRoom)
		p       = &ch.CliProto
	)
	// handshake
	if trd, err = tr.Add(server.Options.HandshakeTimeout, conn); err == nil {
		if key, hb, err = server.authTCP(rr, wr, ch); err == nil {
			tr.Set(trd, hb)
		}
	}
	if err != nil {
		log.Error("handshake failed error(%v)", err)
		if trd != nil {
			tr.Del(trd)
		}
		conn.Close()
		PutBufioReader(rrp, rr)
		PutBufioWriter(wrp, wr)
		return
	}
	// register key->channel
	b = server.Bucket(key)
	b.Put(key, ch)
	// hanshake ok start dispatch goroutine
	go server.dispatchTCP(conn, wrp, wr, ch)
	for {
		// parse request protocol
		if bodyLen, err = server.readTCPRequest(rr, p); err != nil {
			break
		}
		if p.Operation == define.OP_HEARTBEAT {
			// Use a previous timer value if difference between it and a new
			// value is less than TIMER_LAZY_DELAY milliseconds: this allows
			// to minimize the minheap operations for fast connections.
			if !trd.Lazy(hb) {
				tr.Set(trd, hb)
			}
			// heartbeat
			p.Body = nil
			p.Operation = define.OP_HEARTBEAT_REPLY
		} else {
			// process message
			if err = server.operator.Operate(p); err != nil {
				break
			}
		}
		if err = server.writeTCPResponse(ch, wr, p); err != nil {
			break
		}
		if _, err = rr.Discard(bodyLen); err != nil {
			break
		}
	}
	tr.Del(trd)
	conn.Close()
	PutBufioReader(rrp, rr)
	b.Del(key)
	ch.Close()
	if err = server.operator.Disconnect(key, ch.RoomId); err != nil {
		log.Error("%s operator do disconnect error(%v)", key, err)
	}
	if Debug {
		log.Debug("%s serverconn goroutine exit", key)
	}
	return
}

// dispatch accepts connections on the listener and serves requests
// for each incoming connection.  dispatch blocks; the caller typically
// invokes it in a go statement.
func (server *Server) dispatchTCP(conn *net.TCPConn, wrp *sync.Pool, wr *bufio.Writer, ch *Channel) {
	var (
		p   *Proto
		err error
	)
	if Debug {
		log.Debug("start dispatch goroutine")
	}
	for {
		if !ch.Ready() {
			goto failed
		}
		// fetch message from svrbox(server send)
		for {
			if p, err = ch.SvrProto.Get(); err != nil {
				// must be empty error
				break
			}
			// just forward the message
			if err = server.writeTCPResponse(ch, wr, p); err != nil {
				goto failed
			}
			ch.SvrProto.GetAdv()
		}
		// only hungry flush response
		if err = wr.Flush(); err != nil {
			goto failed
		}
	}
failed:
	// wake reader up
	conn.Close()
	PutBufioWriter(wrp, wr)
	if Debug {
		log.Debug("dispatch goroutine exit")
	}
	return
}

// auth for goim handshake with client, use rsa & aes.
func (server *Server) authTCP(rr *bufio.Reader, wr *bufio.Writer, ch *Channel) (subKey string, heartbeat time.Duration, err error) {
	var (
		bodyLen int
		p       = &ch.CliProto
	)
	if bodyLen, err = server.readTCPRequest(rr, p); err != nil {
		return
	}
	if p.Operation != define.OP_AUTH {
		log.Warn("auth operation not valid: %d", p.Operation)
		err = ErrOperation
		return
	}
	if subKey, ch.RoomId, heartbeat, err = server.operator.Connect(p); err != nil {
		return
	}
	p.Body = nil
	p.Operation = define.OP_AUTH_REPLY
	if err = server.writeTCPResponse(ch, wr, p); err != nil {
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
func (server *Server) writeTCPResponse(ch *Channel, wr *bufio.Writer, p *Proto) (err error) {
	var packLen int32
	if Debug {
		log.Debug("write proto: %v", p)
	}
	packLen = RawHeaderSize + int32(len(p.Body))
	p.HeaderLen = RawHeaderSize
	ch.SLock.Lock()
	// if no available memory bufio.Writer auth flush response
	binary.BigEndian.PutInt32(ch.Buf[PackOffset:], packLen)
	binary.BigEndian.PutInt16(ch.Buf[HeaderOffset:], p.HeaderLen)
	binary.BigEndian.PutInt16(ch.Buf[VerOffset:], p.Ver)
	binary.BigEndian.PutInt32(ch.Buf[OperationOffset:], p.Operation)
	binary.BigEndian.PutInt32(ch.Buf[SeqIdOffset:], p.SeqId)
	if _, err = wr.Write(ch.Buf[:]); err != nil {
		return
	}
	if p.Body != nil {
		_, err = wr.Write(p.Body)
	}
	ch.SLock.Unlock()
	return
}

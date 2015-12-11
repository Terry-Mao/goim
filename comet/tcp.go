package main

import (
	"bufio"
	log "code.google.com/p/log4go"
	"github.com/Terry-Mao/goim/libs/define"
	"github.com/Terry-Mao/goim/libs/encoding/binary"
	"github.com/Terry-Mao/goim/libs/io/ioutil"
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
		b   *Bucket
		p   *Proto
		key string
		hb  time.Duration // heartbeat
		err error
		trd *TimerData
		ch  = NewChannel(server.Options.CliProto, server.Options.SvrProto, define.NoRoom)
	)
	// auth
	if trd, err = tr.Add(server.Options.Handshake, conn); err != nil {
		log.Error("handshake: timer.Add() error(%v)", err)
	} else {
		if key, hb, err = server.authTCP(rr, wr, ch); err != nil {
			log.Error("handshake: server.auth error(%v)", err)
		}
		//deltimer
		tr.Del(trd)
	}
	// failed
	if err != nil {
		if err = conn.Close(); err != nil {
			log.Error("handshake: conn.Close() error(%v)", err)
		}
		PutBufioReader(rrp, rr)
		PutBufioWriter(wrp, wr)
		return
	}
	// register key->channel
	b = server.Bucket(key)
	b.Put(key, ch)
	// hanshake ok start dispatch goroutine
	go server.dispatchTCP(conn, wrp, wr, ch, hb, tr)
	for {
		// fetch a proto from channel free list
		if p, err = ch.CliProto.Set(); err != nil {
			// if full, simply close connection
			log.Error("%s fetch client proto error(%v)", key, err)
			goto failed
		}
		// parse request protocol
		if err = server.readTCPRequest(rr, p); err != nil {
			log.Error("%s read client request error(%v)", key, err)
			goto failed
		}
		// send to writer
		ch.CliProto.SetAdv()
		ch.Signal()
	}
failed:
	// dialog finish
	// may call twice
	if err = conn.Close(); err != nil {
		log.Error("reader: conn.Close() error(%v)", err)
	}
	PutBufioReader(rrp, rr)
	b.Del(key)
	if Debug {
		log.Debug("wake up dispatch goroutine")
	}
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
func (server *Server) dispatchTCP(conn *net.TCPConn, wrp *sync.Pool, wr *bufio.Writer, ch *Channel, hb time.Duration, tr *Timer) {
	var (
		p   *Proto
		err error
		trd *TimerData
	)
	if Debug {
		log.Debug("start dispatch goroutine")
	}
	if trd, err = tr.Add(hb, conn); err != nil {
		log.Error("dispatch: timer.Add() error(%v)", err)
		goto failed
	}
	for {
		if !ch.Ready() {
			goto failed
		}
		// fetch message from clibox(client send)
		for {
			if p, err = ch.CliProto.Get(); err != nil {
				// must be empty error
				break
			}
			if p.Operation == define.OP_HEARTBEAT {
				// Use a previous timer value if difference between it and a new
				// value is less than TIMER_LAZY_DELAY milliseconds: this allows
				// to minimize the minheap operations for fast connections.
				if !trd.Lazy(hb) {
					tr.Del(trd)
					if trd, err = tr.Add(hb, conn); err != nil {
						log.Error("dispatch: timer.Add() error(%v)", err)
						goto failed
					}
				}
				// heartbeat
				p.Body = nil
				p.Operation = define.OP_HEARTBEAT_REPLY
			} else {
				// process message
				if err = server.operator.Operate(p); err != nil {
					log.Error("operator.Operate() error(%v)", err)
					goto failed
				}
			}
			if err = server.writeTCPResponse(wr, p); err != nil {
				log.Error("server.writeTCPResponse() error(%v)", err)
				goto failed
			}
			ch.CliProto.GetAdv()
		}
		// fetch message from svrbox(server send)
		for {
			if p, err = ch.SvrProto.Get(); err != nil {
				// must be empty error
				break
			}
			// just forward the message
			if err = server.writeTCPResponse(wr, p); err != nil {
				log.Error("server.writeTCPResponse() error(%v)", err)
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
	if err = conn.Close(); err != nil {
		log.Warn("conn.Close() error(%v)", err)
	}
	// deltimer
	tr.Del(trd)
	PutBufioWriter(wrp, wr)
	if Debug {
		log.Debug("dispatch goroutine exit")
	}
	return
}

// auth for goim handshake with client, use rsa & aes.
func (server *Server) authTCP(rr *bufio.Reader, wr *bufio.Writer, ch *Channel) (subKey string, heartbeat time.Duration, err error) {
	var p *Proto
	// WARN
	// don't adv the cli proto, after auth simply discard it.
	if p, err = ch.CliProto.Set(); err != nil {
		return
	}
	if err = server.readTCPRequest(rr, p); err != nil {
		return
	}
	if p.Operation != define.OP_AUTH {
		log.Warn("auth operation not valid: %d", p.Operation)
		err = ErrOperation
		return
	}
	if subKey, ch.RoomId, heartbeat, err = server.operator.Connect(p); err != nil {
		log.Error("operator.Connect error(%v)", err)
		return
	}
	p.Body = nil
	p.Operation = define.OP_AUTH_REPLY
	if err = server.writeTCPResponse(wr, p); err != nil {
		log.Error("[%s] server.sendTCPResponse() error(%v)", subKey, err)
	}
	err = wr.Flush()
	return
}

// readRequest
func (server *Server) readTCPRequest(rr *bufio.Reader, p *Proto) (err error) {
	var (
		bodySize int32
		buf      []byte
	)
	if buf, err = rr.Peek(RawHeaderSize); err != nil {
		return
	}
	p.PackLen = binary.BigEndian.Int32(buf[PackOffset:HeaderOffset])
	p.HeaderLen = binary.BigEndian.Int16(buf[HeaderOffset:VerOffset])
	p.Ver = binary.BigEndian.Int16(buf[VerOffset:OperationOffset])
	p.Operation = binary.BigEndian.Int32(buf[OperationOffset:SeqIdOffset])
	p.SeqId = binary.BigEndian.Int32(buf[SeqIdOffset:EndOffset])
	if _, err = rr.Discard(RawHeaderSize); err != nil {
		return
	}
	if p.PackLen > MaxPackSize {
		return ErrProtoPackLen
	}
	if p.HeaderLen != RawHeaderSize {
		return ErrProtoHeaderLen
	}
	if bodySize = p.PackLen - int32(p.HeaderLen); bodySize > 0 {
		p.Body = p.Readbuf[0:bodySize]
		err = ioutil.ReadAll(rr, p.Body)
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
	if Debug {
		log.Debug("write proto: %v", p)
	}
	p.PackLen = RawHeaderSize + int32(len(p.Body))
	p.HeaderLen = RawHeaderSize
	// if no available memory bufio.Writer auth flush response
	binary.BigEndian.PutInt32(p.Writebuf[PackOffset:], p.PackLen)
	binary.BigEndian.PutInt16(p.Writebuf[HeaderOffset:], p.HeaderLen)
	binary.BigEndian.PutInt16(p.Writebuf[VerOffset:], p.Ver)
	binary.BigEndian.PutInt32(p.Writebuf[OperationOffset:], p.Operation)
	binary.BigEndian.PutInt32(p.Writebuf[SeqIdOffset:], p.SeqId)
	if _, err = wr.Write(p.Writebuf[:]); err != nil {
		return
	}
	if p.Body != nil {
		_, err = wr.Write(p.Body)
	}
	return
}

package main

import (
	"bufio"
	log "code.google.com/p/log4go"
	"github.com/Terry-Mao/goim/define"
	"github.com/Terry-Mao/goim/libs/io/ioutil"
	"net"
	"sync"
	"time"
)

const (
	maxPackLen    = 1 << 10
	rawHeaderLen  = int16(16)
	packLenSize   = 4
	headerLenSize = 2
	maxPackIntBuf = 4
)

// InitTCP listen all tcp.bind and start accept connections.
func InitTCP() (err error) {
	var (
		listener *net.TCPListener
		addr     *net.TCPAddr
	)
	for _, bind := range Conf.TCPBind {
		if addr, err = net.ResolveTCPAddr("tcp4", bind); err != nil {
			log.Error("net.ResolveTCPAddr(\"tcp4\", \"%s\") error(%v)", bind, err)
			return
		}
		if listener, err = net.ListenTCP("tcp4", addr); err != nil {
			log.Error("net.ListenTCP(\"tcp4\", \"%s\") error(%v)", bind, err)
			return
		}
		log.Debug("start tcp listen: \"%s\"", bind)
		// split N core accept
		for i := 0; i < Conf.MaxProc; i++ {
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
		if err = conn.SetKeepAlive(Conf.TCPKeepalive); err != nil {
			log.Error("conn.SetKeepAlive() error(%v)", err)
			return
		}
		if err = conn.SetReadBuffer(Conf.TCPSndbuf); err != nil {
			log.Error("conn.SetReadBuffer() error(%v)", err)
			return
		}
		if err = conn.SetWriteBuffer(Conf.TCPRcvbuf); err != nil {
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
		rr = NewBufioReaderSize(rrp, conn, Conf.ReadBufSize)  // reader buf
		wr = NewBufioWriterSize(wrp, conn, Conf.WriteBufSize) // writer buf
		// ip addr
		lAddr = conn.LocalAddr().String()
		rAddr = conn.RemoteAddr().String()
	)
	log.Debug("start tcp serve \"%s\" with \"%s\"", lAddr, rAddr)
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
		ch  = NewChannel(Conf.CliProto, Conf.SvrProto, define.NoRoom)
	)
	// auth
	if trd, err = tr.Add(Conf.HandshakeTimeout, conn); err != nil {
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
	log.Debug("wake up dispatch goroutine")
	ch.Finish()
	if err = server.operator.Disconnect(key, ch.RoomId); err != nil {
		log.Error("%s operator do disconnect error(%v)", key, err)
	}
	log.Debug("%s serverconn goroutine exit", key)
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
	log.Debug("start dispatch goroutine")
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
	}
failed:
	// wake reader up
	if err = conn.Close(); err != nil {
		log.Warn("conn.Close() error(%v)", err)
	}
	// deltimer
	tr.Del(trd)
	PutBufioWriter(wrp, wr)
	log.Debug("dispatch goroutine exit")
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
	return
}

// readRequest
func (server *Server) readTCPRequest(rr *bufio.Reader, proto *Proto) (err error) {
	var (
		packLen   int32
		headerLen int16
		bodyLen   int
	)
	if packLen, err = ioutil.ReadBigEndianInt32(rr); err != nil {
		return
	}
	if Conf.Debug {
		log.Debug("packLen: %d", packLen)
	}
	if packLen > maxPackLen {
		return ErrProtoPackLen
	}
	if headerLen, err = ioutil.ReadBigEndianInt16(rr); err != nil {
		return
	}
	if Conf.Debug {
		log.Debug("headerLen: %d", headerLen)
	}
	if headerLen != rawHeaderLen {
		return ErrProtoHeaderLen
	}
	if proto.Ver, err = ioutil.ReadBigEndianInt16(rr); err != nil {
		return
	}
	if Conf.Debug {
		log.Debug("protoVer: %d", proto.Ver)
	}
	if proto.Operation, err = ioutil.ReadBigEndianInt32(rr); err != nil {
		return
	}
	if Conf.Debug {
		log.Debug("operation: %d", proto.Operation)
	}
	if proto.SeqId, err = ioutil.ReadBigEndianInt32(rr); err != nil {
		return
	}
	if Conf.Debug {
		log.Debug("seqId: %d", proto.SeqId)
	}
	bodyLen = int(packLen - int32(headerLen))
	if Conf.Debug {
		log.Debug("read body len: %d", bodyLen)
	}
	if bodyLen > 0 {
		// TODO reuse buf
		proto.Body = make([]byte, bodyLen)
		if err = ioutil.ReadAll(rr, proto.Body); err != nil {
			log.Error("body: ReadAll() error(%v)", err)
			return
		}
	} else {
		proto.Body = nil
	}
	if Conf.Debug {
		log.Debug("read proto: %v", proto)
	}
	return
}

// sendResponse send resp to client, sendResponse must be goroutine safe.
func (server *Server) writeTCPResponse(wr *bufio.Writer, proto *Proto) (err error) {
	if Conf.Debug {
		log.Debug("write proto: %v", proto)
	}
	if err = ioutil.WriteBigEndianInt32(wr, int32(rawHeaderLen)+int32(len(proto.Body))); err != nil {
		return
	}
	if err = ioutil.WriteBigEndianInt16(wr, rawHeaderLen); err != nil {
		return
	}
	if err = ioutil.WriteBigEndianInt16(wr, proto.Ver); err != nil {
		return
	}
	if err = ioutil.WriteBigEndianInt32(wr, proto.Operation); err != nil {
		return
	}
	if err = ioutil.WriteBigEndianInt32(wr, proto.SeqId); err != nil {
		return
	}
	if proto.Body != nil {
		if _, err = wr.Write(proto.Body); err != nil {
			return
		}
	}
	if err = wr.Flush(); err != nil {
		log.Error("tcp wr.Flush() error(%v)", err)
	}
	proto.Reset()
	return
}

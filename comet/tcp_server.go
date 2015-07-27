package main

import (
	"bufio"
	log "code.google.com/p/log4go"
	"net"
	"sync"
	"time"
)

func (server *Server) serveTCP(conn *net.TCPConn, rrp, wrp *sync.Pool, rr *bufio.Reader, wr *bufio.Writer, tr *Timer) {
	var (
		b   *Bucket
		p   *Proto
		hb  time.Duration // heartbeat
		key string
		err error
		trd *TimerData
		ch  = NewChannel(Conf.CliProto, Conf.SvrProto)
		pb  = make([]byte, maxPackIntBuf)
	)
	// auth
	if trd, err = tr.Add(Conf.HandshakeTimeout, conn); err != nil {
		log.Error("handshake: timer.Add() error(%v)", err)
		goto failed
	}
	key, hb, err = server.authTCP(rr, wr, pb, ch)
	tr.Del(trd)
	if err != nil {
		log.Error("server.authTCP() error(%v)", err)
		goto failed
	}
	// register key->channel
	b = server.Bucket(key)
	b.Put(key, ch)
	// hanshake ok start dispatch goroutine
	go server.dispatchTCP(conn, wrp, wr, ch, hb, tr)
	for {
		// fetch a proto from channel free list
		if p, err = ch.CliProto.Set(); err != nil {
			log.Error("%s fetch client proto error(%v)", key, err)
			goto failed
		}
		// parse request protocol
		if err = server.readTCPRequest(rr, pb, p); err != nil {
			log.Error("%s read client request error(%v)", key, err)
			goto failed
		}
		// send to writer
		ch.CliProto.SetAdv()
		select {
		case ch.Signal <- ProtoReady:
		default:
			log.Warn("%s send a signal, but chan is full just ignore", key)
			break
		}
	}
failed:
	// dialog finish
	// may call twice
	if err = conn.Close(); err != nil {
		log.Error("reader: conn.Close() error(%v)")
	}
	PutBufioReader(rrp, rr)
	if b != nil {
		b.Del(key)
		// don't use close chan, Signal can be reused
		// if chan full, writer goroutine next fetch from chan will exit
		// if chan empty, send a 0(close) let the writer exit
		log.Debug("wake up dispatch goroutine")
		select {
		case ch.Signal <- ProtoFinsh:
		default:
			log.Warn("%s send proto finish signal, but chan is full just ignore", key)
		}
	}
	if err = server.operator.Disconnect(key); err != nil {
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
		sig int
		pb  = make([]byte, maxPackIntBuf) // avoid false sharing
	)
	log.Debug("start dispatch goroutine")
	if trd, err = tr.Add(hb, conn); err != nil {
		log.Error("dispatch: timer.Add() error(%v)", err)
		goto failed
	}
	for {
		if sig = <-ch.Signal; sig == 0 {
			goto failed
		}
		// fetch message from clibox(client send)
		for {
			if p, err = ch.CliProto.Get(); err != nil {
				break
			}
			if p.Operation == OP_HEARTBEAT {
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
				p.Operation = OP_HEARTBEAT_REPLY
			} else {
				// process message
				if err = server.operator.Operate(p); err != nil {
					log.Error("operator.Operate() error(%v)", err)
					goto failed
				}
			}
			if err = server.writeTCPResponse(wr, pb, p); err != nil {
				log.Error("server.sendTCPResponse() error(%v)", err)
				goto failed
			}
			ch.CliProto.GetAdv()
		}
		// fetch message from svrbox(server send)
		for {
			if p, err = ch.SvrProto.Get(); err != nil {
				break
			}
			// just forward the message
			if err = server.writeTCPResponse(wr, pb, p); err != nil {
				log.Error("server.sendTCPResponse() error(%v)", err)
				goto failed
			}
			ch.SvrProto.GetAdv()
		}
	}
failed:
	// wake reader up
	if err = conn.Close(); err != nil {
		log.Error("conn.Close() error(%v)", err)
	}
	// deltimer
	tr.Del(trd)
	PutBufioWriter(wrp, wr)
	log.Debug("dispatch goroutine exit")
	return
}

// auth for goim handshake with client, use rsa & aes.
func (server *Server) authTCP(rr *bufio.Reader, wr *bufio.Writer, pb []byte, ch *Channel) (subKey string, heartbeat time.Duration, err error) {
	var p *Proto
	// WARN
	// don't adv the cli proto, after auth simply discard it.
	if p, err = ch.CliProto.Set(); err != nil {
		return
	}
	if err = server.readTCPRequest(rr, pb, p); err != nil {
		return
	}
	if p.Operation != OP_AUTH {
		log.Warn("auth operation not valid: %d", p.Operation)
		err = ErrOperation
		return
	}
	if subKey, heartbeat, err = server.operator.Connect(p); err != nil {
		log.Error("operator.Connect error(%v)", err)
		return
	}
	p.Body = nil
	p.Operation = OP_AUTH_REPLY
	if err = server.writeTCPResponse(wr, pb, p); err != nil {
		log.Error("[%s] server.sendTCPResponse() error(%v)", subKey, err)
	}
	return
}

// readRequest
func (server *Server) readTCPRequest(rr *bufio.Reader, pb []byte, proto *Proto) (err error) {
	var (
		packLen   int32
		headerLen int16
		bodyLen   int
	)
	if err = ReadAll(rr, pb[:packLenSize]); err != nil {
		return
	}
	packLen = BigEndian.Int32(pb[:packLenSize])
	log.Debug("packLen: %d", packLen)
	if packLen > maxPackLen {
		return ErrProtoPackLen
	}
	if err = ReadAll(rr, pb[:headerLenSize]); err != nil {
		return
	}
	headerLen = BigEndian.Int16(pb[:headerLenSize])
	log.Debug("headerLen: %d", headerLen)
	if headerLen != rawHeaderLen {
		return ErrProtoHeaderLen
	}
	if err = ReadAll(rr, pb[:VerSize]); err != nil {
		return
	}
	proto.Ver = BigEndian.Int16(pb[:VerSize])
	log.Debug("protoVer: %d", proto.Ver)
	if err = ReadAll(rr, pb[:OperationSize]); err != nil {
		return
	}
	proto.Operation = BigEndian.Int32(pb[:OperationSize])
	log.Debug("operation: %d", proto.Operation)
	if err = ReadAll(rr, pb[:SeqIdSize]); err != nil {
		return
	}
	proto.SeqId = BigEndian.Int32(pb[:SeqIdSize])
	log.Debug("seqId: %d", proto.SeqId)
	bodyLen = int(packLen - int32(headerLen))
	log.Debug("read body len: %d", bodyLen)
	if bodyLen > 0 {
		proto.Body = make([]byte, bodyLen)
		if err = ReadAll(rr, proto.Body); err != nil {
			log.Error("body: ReadAll() error(%v)", err)
			return
		}
	} else {
		proto.Body = nil
	}
	log.Debug("read proto: %s", proto)
	return
}

// sendResponse send resp to client, sendResponse must be goroutine safe.
func (server *Server) writeTCPResponse(wr *bufio.Writer, pb []byte, proto *Proto) (err error) {
	log.Debug("write proto: %s", proto)
	BigEndian.PutInt32(pb[:packLenSize], int32(rawHeaderLen)+int32(len(proto.Body)))
	if _, err = wr.Write(pb[:packLenSize]); err != nil {
		return
	}
	BigEndian.PutInt16(pb[:headerLenSize], rawHeaderLen)
	if _, err = wr.Write(pb[:headerLenSize]); err != nil {
		return
	}
	BigEndian.PutInt16(pb[:VerSize], proto.Ver)
	if _, err = wr.Write(pb[:VerSize]); err != nil {
		return
	}
	BigEndian.PutInt32(pb[:OperationSize], proto.Operation)
	if _, err = wr.Write(pb[:OperationSize]); err != nil {
		return
	}
	BigEndian.PutInt32(pb[:SeqIdSize], proto.SeqId)
	if _, err = wr.Write(pb[:SeqIdSize]); err != nil {
		return
	}
	if proto.Body != nil {
		if _, err = wr.Write(proto.Body); err != nil {
			return
		}
	}
	return wr.Flush()
}

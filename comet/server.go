package main

import (
	log "code.google.com/p/log4go"
	"crypto/cipher"
	"github.com/Terry-Mao/goim/libs/hash/cityhash"
	"io"
	"time"
)

var (
	maxInt = 1<<31 - 1
)

type Server struct {
	buckets   []*Bucket // subkey bucket
	bucketIdx uint32
	round     *Round // accept round store
	codec     ServerCodec
	operator  Operator
	cryptor   Cryptor
}

// NewServer returns a new Server.
func NewServer(b []*Bucket, r *Round, c ServerCodec, o Operator, d Cryptor) *Server {
	s := new(Server)
	log.Debug("server: use default server codec")
	s.buckets = b
	s.bucketIdx = uint32(len(b))
	s.round = r
	s.codec = c
	s.operator = o
	s.cryptor = d
	return s
}

func (server *Server) serve(rr io.Reader, wr io.Writer, fr Flusher, cr io.Closer, r int) {
	var (
		b   *Bucket
		ch  *Channel
		hb  time.Duration // heartbeat
		key string
		ebm cipher.BlockMode
		dbm cipher.BlockMode
		err error
		trd *TimerData
		p   = new(Proto)
		tr  = server.round.Timer(r)
		s   = server.round.Session(r)
	)
	// handshake & auth
	if trd, err = tr.Add(Conf.HandshakeTimeout, cr); err != nil {
		log.Error("handshake: timer.Add() error(%v)", err)
	} else {
		if ebm, dbm, err = server.handshake(rr, wr, fr, s, p); err != nil {
			log.Error("handshake: server.handshake error(%v)", err)
		} else {
			if key, hb, err = server.auth(rr, wr, fr, dbm, p); err != nil {
				log.Error("handshake: server.auth error(%v)", err)
			}
		}
		//deltimer
		tr.Del(trd)
	}
	// failed
	if err != nil {
		if err = cr.Close(); err != nil {
			log.Error("handshake: fr.Close() error(%v)", err)
		}
		return
	}
	// TODO how to reuse channel
	// register key->channel
	b = server.Bucket(key)
	ch = NewChannel(Conf.CliProto, Conf.SvrProto)
	b.Put(key, ch)
	// hanshake ok start dispatch goroutine
	log.Debug("%s serverconn goroutine start", key)
	go server.dispatch(cr, wr, fr, ch, ebm, dbm, hb, tr)
	for {
		// fetch a proto from channel free list
		if p, err = ch.CliProto.Set(); err != nil {
			log.Error("%s fetch client proto error(%v)", key, err)
			break
		}
		// parse request protocol
		if err = server.readRequest(rr, p); err != nil {
			log.Error("%s read client request error(%v)", key, err)
			break
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
	// dialog finish
	// revoke the subkey
	// revoke the remote subkey
	// close the net.Conn
	// read & write goroutine
	// return channel to bucket's free list
	b.Del(key)
	if err = server.operator.Disconnect(key); err != nil {
		log.Error("%s operator do disconnect error(%v)", key, err)
	}
	// may call twice
	if err = cr.Close(); err != nil {
		log.Error("reader: cr.Close() error(%v)")
	}
	// don't use close chan, Signal can be reused
	// if chan full, writer goroutine next fetch from chan will exit
	// if chan empty, send a 0(close) let the writer exit
	log.Debug("wake up dispatch goroutine")
	select {
	case ch.Signal <- ProtoFinsh:
	default:
		log.Warn("%s send proto finish signal, but chan is full just ignore", key)
	}
	log.Debug("%s serverconn goroutine exit", key)
	return
}

// dispatch accepts connections on the listener and serves requests
// for each incoming connection.  dispatch blocks; the caller typically
// invokes it in a go statement.
func (server *Server) dispatch(cr io.Closer, wr io.Writer, fr Flusher, ch *Channel, ebm, dbm cipher.BlockMode, hb time.Duration, tr *Timer) {
	var (
		p   *Proto
		err error
		trd *TimerData
		sig int
	)
	log.Debug("start dispatch goroutine")
	if trd, err = tr.Add(hb, cr); err != nil {
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
				log.Debug("channel no more client message, wait signal")
				break
			}
			if p.Operation == OP_HEARTBEAT {
				// Use a previous timer value if difference between it and a new
				// value is less than TIMER_LAZY_DELAY milliseconds: this allows
				// to minimize the minheap operations for fast connections.
				if !trd.Lazy(hb) {
					tr.Del(trd)
					if trd, err = tr.Add(hb, cr); err != nil {
						log.Error("dispatch: timer.Add() error(%v)", err)
						goto failed
					}
				}
				// heartbeat
				p.Body = nil
				p.Operation = OP_HEARTBEAT_REPLY
				log.Debug("heartbeat proto: %v", p)
			} else {
				// aes decrypt body
				if p.Body, err = server.cryptor.Decrypt(dbm, p.Body); err != nil {
					log.Error("decrypt client proto error(%v)", err)
					goto failed
				}
				// process message
				if err = server.operator.Operate(p); err != nil {
					log.Error("operator.Operate() error(%v)", err)
					goto failed
				}
				if p.Body, err = server.cryptor.Encrypt(ebm, p.Body); err != nil {
					log.Error("encrypt client proto error(%v)", err)
					goto failed
				}
			}
			if err = server.sendResponse(wr, fr, p); err != nil {
				log.Error("server.SendResponse() error(%v)", err)
				goto failed
			}
			ch.CliProto.GetAdv()
		}
		// fetch message from svrbox(server send)
		for {
			if p, err = ch.SvrProto.Get(); err != nil {
				log.Debug("channel no more server message, wait signal")
				break
			}
			if p.Body, err = server.cryptor.Encrypt(ebm, p.Body); err != nil {
				log.Error("encrypt server proto error(%v)", err)
				goto failed
			}
			// just forward the message
			if err = server.sendResponse(wr, fr, p); err != nil {
				log.Error("server.SendResponse() error(%v)", err)
				goto failed
			}
			ch.SvrProto.GetAdv()
		}
	}
failed:
	// wake reader up
	if err = cr.Close(); err != nil {
		log.Error("cr.Close() error(%v)", err)
	}
	// deltimer
	tr.Del(trd)
	log.Debug("dispatch goroutine exit")
	return
}

// handshake for goim handshake with client, use rsa & aes.
func (server *Server) handshake(rr io.Reader, wr io.Writer, fr Flusher, s *Session, p *Proto) (ebm, dbm cipher.BlockMode, err error) {
	var (
		ki, sidBytes []byte
	)
	// exchange key
	log.Debug("get handshake request protocol")
	if err = server.readRequest(rr, p); err != nil {
		return
	}
	if p.Operation == OP_HANDSHAKE {
		log.Debug("handshake cipher body : %v", p.Body)
		if ki, err = server.cryptor.Exchange(RSAPri, p.Body); err != nil {
			log.Error("server.cryptor.Exchange() error(%v)", err)
			return
		}
		p.Operation = OP_HANDSHAKE_REPLY
		sidBytes = []byte(s.Put(ki, SessionExpire))
		log.Debug("session id: \"%s\"", sidBytes)
	} else if p.Operation == OP_HANDSHAKE_SID {
		// raw message for find session -> aes key & iv
		// if sniffed by somebody, it's safe here
		if ki = s.Get(string(p.Body)); ki == nil {
			log.Error("session.Get(\"%s\") not exists", string(p.Body))
			return
		}
		p.Operation = OP_HANDSHAKE_SID_REPLY
	} else {
		log.Warn("handshake operation not valid: %d", p.Operation)
		err = ErrOperation
		return
	}
	if ebm, dbm, err = server.cryptor.Cryptor(ki); err != nil {
		log.Error("server.cryptor.Cryptor() error(%v)", err)
		return
	}
	if p.Body, err = server.cryptor.Encrypt(ebm, sidBytes); err != nil {
		log.Error("server.cryptor.Encrypt() sessionid error(%v)", err)
		return
	}
	log.Debug("send handshake response protocol")
	if err = server.sendResponse(wr, fr, p); err != nil {
		log.Error("handshake reply server.SendResponse() error(%v)", err)
	}
	return
}

// auth for goim handshake with client, use rsa & aes.
func (server *Server) auth(rr io.Reader, wr io.Writer, fr Flusher, dbm cipher.BlockMode, p *Proto) (subKey string, heartbeat time.Duration, err error) {
	log.Debug("get auth request protocol")
	if err = server.readRequest(rr, p); err != nil {
		return
	}
	if p.Operation != OP_AUTH {
		log.Warn("auth operation not valid: %d", p.Operation)
		err = ErrOperation
		return
	}
	if p.Body, err = server.cryptor.Decrypt(dbm, p.Body); err != nil {
		log.Error("auth decrypt client proto error(%v)", err)
		return
	}
	if subKey, heartbeat, err = server.operator.Connect(p); err != nil {
		log.Error("operator.Connect error(%v)", err)
		return
	}
	log.Debug("send auth response protocol")
	p.Body = nil
	p.Operation = OP_AUTH_REPLY
	if err = server.sendResponse(wr, fr, p); err != nil {
		log.Error("[%s] server.SendResponse() error(%v)", subKey, err)
	}
	return
}

// readRequest
func (server *Server) readRequest(rr io.Reader, proto *Proto) (err error) {
	if err = server.readRequestHeader(rr, proto); err != nil {
		return
	}
	// read body
	if err = server.codec.ReadRequestBody(rr, proto); err != nil {
	}
	log.Debug("read request finish, proto: %v", proto)
	return
}

func (server *Server) readRequestHeader(rr io.Reader, proto *Proto) (err error) {
	if err = server.codec.ReadRequestHeader(rr, proto); err != nil {
		log.Error("codec.ReadRequestHeader() error(%v)", err)
	}
	return
}

func (server *Server) readRequestBody(rr io.Reader, proto *Proto) (err error) {
	if err = server.codec.ReadRequestHeader(rr, proto); err != nil {
		log.Error("codec.ReadRequestHeader() error(%v)", err)
	}
	return
}

// sendResponse send resp to client, sendResponse must be goroutine safe.
func (server *Server) sendResponse(wr io.Writer, fr Flusher, proto *Proto) (err error) {
	// do not set write timer, if pendding no heartbeat will receive, then conn.Close() will wake this up.
	if err = server.codec.WriteResponse(wr, fr, proto); err != nil {
		log.Error("server.codec.WriteResponse() error(%v)", err)
	}
	return
}

func (server *Server) Bucket(subKey string) *Bucket {
	idx := cityhash.CityHash32([]byte(subKey), uint32(len(subKey))) % server.bucketIdx
	log.Debug("\"%s\" hit channel bucket index: %d use cityhash", subKey, idx)
	return server.buckets[idx]
}

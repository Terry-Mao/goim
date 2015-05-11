package main

import (
	"bufio"
	log "code.google.com/p/log4go"
	"crypto/aes"
	"crypto/cipher"
	"github.com/Terry-Mao/goim/libs/crypto/rsa"
	"net"
	"sync"
	"time"
)

var (
	defaultOperator = new(IMOperator)
	aesKeyLen       = 16
	maxInt          = 1<<31 - 1
)

type ServerCodec interface {
	ReadRequestHeader(*bufio.Reader, *Proto) error
	ReadRequestBody(*bufio.Reader, *Proto) error
	// WriteResponse must be safe for concurrent use by multiple goroutines.
	WriteResponse(*bufio.Writer, *Proto) error
}

type Server struct {
	buckets []*Bucket // subkey bucket
	round   *Round    // accept round store
	codec   ServerCodec
}

// NewServer returns a new Server.
func NewServer() *Server {
	s := new(Server)
	log.Debug("server: use default server codec")
	s.codec = new(DefaultServerCodec)
	s.buckets = make([]*Bucket, Conf.Bucket)
	for i := 0; i < Conf.Bucket; i++ {
		s.buckets[i] = NewBucket(Conf.Channel, Conf.CliProto, Conf.SvrProto)
	}
	s.round = NewRound(Conf.ReadBuf, Conf.WriteBuf, Conf.Timer, Conf.TimerSize, Conf.HandshakeProto, Conf.HandshakeProtoSize)
	return s
}

// Accept accepts connections on the listener and serves requests
// for each incoming connection.  Accept blocks; the caller typically
// invokes it in a go statement.
func (server *Server) Accept(lis net.Listener, i int) {
	var (
		conn net.Conn
		err  error
	)
	for {
		log.Debug("server: accept round: %d", i)
		if conn, err = lis.Accept(); err != nil {
			// if listener close then return
			log.Error("listener.Accept(\"%s\") error(%v)", lis.Addr().String(), err)
			return
		}
		go server.serveConn(conn, i)
		if i++; i == maxInt {
			i = 0
		}
	}
}

func (server *Server) serveConn(conn net.Conn, r int) {
	var (
		block     cipher.Block
		subKey    string
		err       error
		heartbeat time.Duration
		bucket    *Bucket
		channel   *Channel
		timerd    *TimerData
		// timer
		timer = server.round.Timer(r)
		// bufpool
		rp = server.round.Reader(r) // reader
		wp = server.round.Writer(r) // writer
		// free proto
		fp = server.round.Proto(r)
		// bufio
		rd = NewBufioReaderSize(rp, conn, Conf.ReadBufSize)  // read buf
		wr = NewBufioWriterSize(wp, conn, Conf.WriteBufSize) // write buf
		// proto
		proto = fp.Get()
		// ip addr
		lAddr = conn.LocalAddr().String()
		rAddr = conn.RemoteAddr().String()
	)
	// handshake
	if timerd, err = timer.Add(Conf.HandshakeTimeout, conn); err != nil {
		log.Error("\"%s\" handshake timer.Add() error(%v)", rAddr, err)
	} else {
		log.Debug("handshake \"%s\" with \"%s\"", lAddr, rAddr)
		if block, subKey, heartbeat, bucket, channel, err = server.handshake(rd, wr, proto); err != nil {
			log.Error("\"%s\"->\"%s\" handshake() error(%v)", lAddr, rAddr, err)
		}
		//deltimer
		timer.Del(timerd)
	}
	fp.Free(proto)
	// failed
	if err != nil {
		if err = conn.Close(); err != nil {
			log.Error("conn.Close(\"%s\", \"%s\") error(%v)", lAddr, rAddr, err)
		}
		PutBufioReader(rp, rd)
		PutBufioWriter(wp, wr)
		return
	}
	// hanshake ok start dispatch goroutine
	log.Debug("%s[%s] serverconn goroutine start", subKey, rAddr)
	go server.dispatch(conn, wr, wp, channel, block, heartbeat, timer, rAddr)
	for {
		// fetch a proto from channel free list
		if proto, err = channel.CliProto.Set(); err != nil {
			log.Error("%s[%s] fetch client proto error(%v)", subKey, rAddr, err)
			break
		}
		// parse request protocol
		if err = server.readRequest(rd, proto); err != nil {
			log.Error("%s[%s] read client request error(%v)", subKey, rAddr, err)
			break
		}
		// send to writer
		channel.CliProto.SetAdv()
		select {
		case channel.Signal <- ProtoReady:
		default:
			log.Warn("%s[%s] send a signal, but chan is full just ignore", subKey, rAddr)
			break
		}
	}
	// dialog finish
	// put back reader buf
	// revoke the subkey
	// revoke the remote subkey
	// close the net.Conn
	// read & write goroutine
	// return channel to bucket's free list
	PutBufioReader(rp, rd)
	bucket.Del(subKey)
	if err = defaultOperator.Disconnect(subKey); err != nil {
		log.Error("%s[%s] operator do disconnect error(%v)", subKey, rAddr, err)
	}
	// may call twice
	if err = conn.Close(); err != nil {
		log.Error("conn.Close(\"%s\", \"%s\") error(%v)", lAddr, rAddr, err)
	}
	// don't use close chan, Signal can be reused
	// if chan full, writer goroutine next fetch from chan will exit
	// if chan empty, send a 0(close) let the writer exit
	log.Debug("wake up dispatch goroutine")
	select {
	case channel.Signal <- ProtoFinsh:
	default:
		log.Warn("%s[%s] send proto finish signal, but chan is full just ignore", subKey, rAddr)
	}
	log.Debug("%s[%s] serverconn goroutine exit", subKey, rAddr)
}

// dispatch accepts connections on the listener and serves requests
// for each incoming connection.  dispatch blocks; the caller typically
// invokes it in a go statement.
func (server *Server) dispatch(conn net.Conn, wr *bufio.Writer, wp *sync.Pool, channel *Channel, block cipher.Block, heartbeat time.Duration, timer *Timer, rAddr string) {
	var (
		err    error
		proto  *Proto
		signal int
		timerd *TimerData
	)
	log.Debug("\"%s\" start dispatch goroutine", rAddr)
	log.Debug("\"%s\" first set heartbeat timer", rAddr)
	if timerd, err = timer.Add(heartbeat, conn); err != nil {
		log.Error("\"%s\" dispatch timer.Add() error(%v)", rAddr, err)
		goto failed
	}
	for {
		if signal = <-channel.Signal; signal == 0 {
			goto failed
		}
		// fetch message from clibox(client send)
		for {
			if proto, err = channel.CliProto.Get(); err != nil {
				log.Debug("\"%s\" channel no more client message, wait signal", rAddr)
				break
			}
			if proto.Operation == OP_HEARTBEAT {
				// Use a previous timer value if difference between it and a new
				// value is less than TIMER_LAZY_DELAY milliseconds: this allows
				// to minimize the minheap operations for fast connections.
				if !timerd.Lazy(heartbeat) {
					timer.Del(timerd)
					if timerd, err = timer.Add(heartbeat, conn); err != nil {
						log.Error("\"%s\" dispatch timer.Add() error(%v)", rAddr, err)
						goto failed
					}
				}
				// heartbeat
				proto.Body = nil
				proto.Operation = OP_HEARTBEAT_REPLY
				log.Debug("\"%s\" heartbeat proto: %v", rAddr, proto)
			} else {
				// aes decrypt body
				if err = proto.Decrypt(block); err != nil {
					log.Error("\"%s\" decrypt client proto error(%v)", rAddr, err)
					goto failed
				}
				// process message
				if err = defaultOperator.Operate(proto); err != nil {
					log.Error("\"%s\" operator.Operate() error(%v)", rAddr, err)
					goto failed
				}
				if err = proto.Encrypt(block); err != nil {
					log.Error("\"%s\" encrypt client proto error(%v)", rAddr, err)
					goto failed
				}
			}
			if err = server.sendResponse(wr, proto); err != nil {
				log.Error("\"%s\" server.SendResponse() error(%v)", rAddr, err)
				goto failed
			}
			channel.CliProto.GetAdv()
		}
		// fetch message from svrbox(server send)
		for {
			if proto, err = channel.SvrProto.Get(); err != nil {
				log.Debug("\"%s\" channel no more server message, wait signal", rAddr)
				break
			}
			if err = proto.Encrypt(block); err != nil {
				log.Error("\"%s\" encrypt server proto error(%v)", rAddr, err)
				goto failed
			}
			// just forward the message
			if err = server.sendResponse(wr, proto); err != nil {
				log.Error("\"%s\" server.SendResponse() error(%v)", rAddr, err)
				goto failed
			}
			channel.SvrProto.GetAdv()
		}
	}
failed:
	// wake reader up
	//  put writer buf & back encrypter buf & decrypter buf
	PutBufioWriter(wp, wr)
	if err = conn.Close(); err != nil {
		log.Error("conn.Close(\"%s\") error(%v)", rAddr, err)
	}
	// deltimer
	timer.Del(timerd)
	log.Debug("\"%s\" dispatch goroutine exit", rAddr)
	return
}

// handshake for goim handshake with client, use rsa & aes.
func (server *Server) handshake(rd *bufio.Reader, wr *bufio.Writer, proto *Proto) (block cipher.Block, subKey string, heartbeat time.Duration, bucket *Bucket, channel *Channel, err error) {
	var (
		aesKey []byte
	)
	// 1. exchange aes key
	log.Debug("get handshake request protocol")
	if err = server.readRequest(rd, proto); err != nil {
		return
	}
	if proto.Operation != OP_HANDSHAKE {
		log.Warn("handshake operation not valid: %d", proto.Operation)
		err = ErrOperation
		return
	}
	log.Debug("handshake cipher body : %v", proto.Body)
	if aesKey, err = rsa.Decrypt(proto.Body, RSAPri); err != nil {
		log.Error("rsa.Decrypt() error(%v)", err)
		return
	}
	log.Debug("handshake aesKey : 0x%x", aesKey)
	if len(aesKey) != aesKeyLen {
		log.Warn("handshake aes key size not valid: %d", len(aesKey))
		err = ErrHandshake
		return
	}
	if block, err = aes.NewCipher(aesKey); err != nil {
		log.Error("handshake aes.NewCipher() error(%v)", err)
		return
	}
	log.Debug("send handshake response protocol")
	proto.Body = nil
	proto.Operation = OP_HANDSHAKE_REPLY
	if err = server.sendResponse(wr, proto); err != nil {
		log.Error("handshake reply server.SendResponse() error(%v)", err)
		return
	}
	// 2. auth token
	log.Debug("get auth request protocol")
	if err = server.readRequest(rd, proto); err != nil {
		return
	}
	if proto.Operation != OP_AUTH {
		log.Warn("auth operation not valid: %d", proto.Operation)
		err = ErrOperation
		return
	}
	if err = proto.Decrypt(block); err != nil {
		log.Error("auth decrypt client proto error(%v)", err)
		return
	}
	if subKey, heartbeat, err = defaultOperator.Connect(proto); err != nil {
		log.Error("operator.Connect error(%v)", err)
		return
	}
	// TODO how to reuse channel
	// register key->channel
	bucket = server.Bucket(subKey)
	channel = NewChannel(Conf.CliProto, Conf.SvrProto)
	bucket.Put(subKey, channel)
	log.Debug("send auth response protocol")
	proto.Body = nil
	proto.Operation = OP_AUTH_REPLY
	if err = server.sendResponse(wr, proto); err != nil {
		log.Error("[%s] server.SendResponse() error(%v)", subKey, err)
	}
	return
}

// readRequest
func (server *Server) readRequest(rd *bufio.Reader, proto *Proto) (err error) {
	if err = server.readRequestHeader(rd, proto); err != nil {
		return
	}
	// read body
	if err = server.codec.ReadRequestBody(rd, proto); err != nil {
	}
	log.Debug("read request finish, proto: %v", proto)
	return
}

func (server *Server) readRequestHeader(rd *bufio.Reader, proto *Proto) (err error) {
	if err = server.codec.ReadRequestHeader(rd, proto); err != nil {
		log.Error("codec.ReadRequestHeader() error(%v)", err)
	}
	return
}

func (server *Server) readRequestBody(rd *bufio.Reader, proto *Proto) (err error) {
	if err = server.codec.ReadRequestHeader(rd, proto); err != nil {
		log.Error("codec.ReadRequestHeader() error(%v)", err)
	}
	return
}

// sendResponse send resp to client, sendResponse must be goroutine safe.
func (server *Server) sendResponse(wr *bufio.Writer, proto *Proto) (err error) {
	// do not set write timer, if pendding no heartbeat will receive, then conn.Close() will wake this up.
	if err = server.codec.WriteResponse(wr, proto); err != nil {
		log.Error("server.codec.WriteResponse() error(%v)", err)
	}
	return
}

func (server *Server) Bucket(subKey string) *Bucket {
	h := NewMurmur3C()
	h.Write([]byte(subKey))
	idx := h.Sum32() & uint32(Conf.Bucket-1)
	log.Debug("\"%s\" hit channel bucket index: %d", subKey, idx)
	return server.buckets[idx]
}

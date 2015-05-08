package main

import (
	"bufio"
	log "code.google.com/p/log4go"
	"github.com/Terry-Mao/goim/libs/crypto/aes"
	"github.com/Terry-Mao/goim/libs/crypto/padding"
	"github.com/Terry-Mao/goim/libs/crypto/rsa"
	"net"
	"sync"
	"time"
)

var (
	defaultOperator = new(IMOperator)
	aesKeyLen       = 16
	maxInt          = 2 ^ 31 - 1
)

// Proto is a request&response written before every goim connect.  It is used internally
// but documented here as an aid to debugging, such as when analyzing
// network traffic.
type Proto struct {
	PackLen   int32  // package length
	HeaderLen int16  // header length
	Ver       int16  // protocol version
	Operation int32  // operation for request
	SeqId     int32  // sequence number chosen by client
	Body      []byte // body
}

type ServerCodec interface {
	ReadRequestHeader(*bufio.Reader, *Proto) error
	ReadRequestBody(*bufio.Reader, *Proto) error
	// WriteResponse must be safe for concurrent use by multiple goroutines.
	WriteResponse(*bufio.Writer, *Proto) error
}

type Server struct {
	buckets []*Bucket
	rPool   []*sync.Pool
	wPool   []*sync.Pool
	timers  []*Timer
	codec   ServerCodec
}

// NewServer returns a new Server.
func NewServer() *Server {
	s := new(Server)
	log.Debug("server: create %d bucket for store sub channel", Conf.Bucket)
	s.buckets = make([]*Bucket, Conf.Bucket)
	for i := 0; i < Conf.Bucket; i++ {
		s.buckets[i] = NewBucket(Conf.Channel, Conf.CliProto, Conf.SvrProto)
	}
	log.Debug("server: create %d reader buffer pool", Conf.ReadBuf)
	s.rPool = make([]*sync.Pool, Conf.ReadBuf)
	for i := 0; i < Conf.ReadBuf; i++ {
		s.rPool[i] = new(sync.Pool)
	}
	log.Debug("server: create %d writer buffer pool", Conf.WriteBuf)
	s.wPool = make([]*sync.Pool, Conf.WriteBuf)
	for i := 0; i < Conf.WriteBuf; i++ {
		s.wPool[i] = new(sync.Pool)
	}
	log.Debug("server: create %d time", Conf.Timer)
	s.timers = make([]*Timer, Conf.Timer)
	for i := 0; i < Conf.Timer; i++ {
		s.timers[i] = NewTimer(Conf.TimerSize)
	}
	log.Debug("server: use default server codec")
	s.codec = new(DefaultServerCodec)
	return s
}

// Accept accepts connections on the listener and serves requests
// for each incoming connection.  Accept blocks; the caller typically
// invokes it in a go statement.
func (server *Server) Accept(lis net.Listener) {
	var (
		i    = 0
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
		aesKey    []byte
		subKey    string
		err       error
		heartbeat time.Duration
		bucket    *Bucket
		channel   *Channel
		// proto
		proto *Proto
		// timer
		timer = server.timers[r&(Conf.Timer-1)]
		// bufio&bufpool
		rp = server.rPool[r&(Conf.ReadBuf-1)]
		wp = server.wPool[r&(Conf.WriteBuf-1)]
		rd = newBufioReaderSize(rp, conn, Conf.ReadBufSize)
		wr = newBufioWriterSize(wp, conn, Conf.WriteBufSize)
		// ip addr
		lAddr = conn.LocalAddr().String()
		rAddr = conn.RemoteAddr().String()
	)
	// handshake
	//timerd.Set(Conf.HandshakeTimeout, conn)
	//if err = timer.Push(timerd); err != nil {
	//	log.Error("\"%s\" handshake timer.Push() error(%v)", rAddr, err)
	//} else {
	log.Debug("handshake \"%s\" with \"%s\"", lAddr, rAddr)
	if aesKey, subKey, heartbeat, bucket, channel, err = server.handshake(rd, wr); err != nil {
		log.Error("\"%s\"->\"%s\" handshake() error(%v)", lAddr, rAddr, err)
	}
	// deltimer
	//	timerd, _ = timer.Remove(timerd)
	//}
	// failed
	if err != nil {
		if err = conn.Close(); err != nil {
			log.Error("conn.Close(\"%s\", \"%s\") error(%v)", lAddr, rAddr, err)
		}
		putBufioReader(rp, rd)
		putBufioWriter(wp, wr)
		return
	}
	// hanshake ok start dispatch goroutine
	log.Debug("%s[%s] serverconn goroutine start", subKey, rAddr)
	log.Debug("[%s] aes key: %v, sub key: \"%s\"", rAddr, aesKey, subKey)
	go server.dispatch(conn, wr, wp, channel, aesKey, heartbeat, timer, rAddr, lAddr)
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
		// aes decrypt body
		if proto.Body != nil {
			if proto.Body, err = aes.ECBDecrypt(proto.Body, aesKey, padding.PKCS7); err != nil {
				log.Error("%s[%s] decrypt client proto error(%v)", subKey, rAddr, err)
				break
			}
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
	putBufioReader(rp, rd)
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
func (server *Server) dispatch(conn net.Conn, wr *bufio.Writer, wp *sync.Pool, channel *Channel, aesKey []byte, heartbeat time.Duration, timer *Timer, rAddr, lAddr string) {
	var (
		err    error
		proto  *Proto
		signal int
	)
	log.Debug("\"%s\" start dispatch goroutine", rAddr)
	log.Debug("\"%s\" first set heartbeat timer", rAddr)
	//timerd.Set(heartbeat, conn)
	//if err = timer.Push(timerd); err != nil {
	//	log.Error("\"%s\" dispatch timer.Update() error(%v)", rAddr, err)
	//	goto failed
	//}
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
				// TODO
				// Use a previous timer value if difference between it and a new
				// value is less than TIMER_LAZY_DELAY milliseconds: this allows
				// to minimize the minheap operations for fast connections.
				// handshake push a timer, reuse
				/*
					if timerd, err = timer.Update(timerd, heartbeat); err != nil {
						log.Error("\"%s\" dispatch timer.Update() error(%v)", rAddr, err)
						goto failed
					}
				*/
				// heartbeat
				proto.Body = nil
				proto.Operation = OP_HEARTBEAT_REPLY
				log.Debug("\"%s\" heartbeat proto: %v", rAddr, proto)
			} else {
				// process message
				if err = defaultOperator.Operate(proto); err != nil {
					log.Error("\"%s\" operator.Operate() error(%v)", rAddr, err)
					goto failed
				}
				if proto.Body != nil {
					if proto.Body, err = aes.ECBEncrypt(proto.Body, aesKey, padding.PKCS7); err != nil {
						log.Error("\"%s\" aes.Encrypt() error(%v)", rAddr, err)
						goto failed
					}
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
			if proto.Body != nil {
				if proto.Body, err = aes.ECBEncrypt(proto.Body, aesKey, padding.PKCS7); err != nil {
					log.Error("\"%s\" aes.Encrypt() error(%v)", rAddr, err)
					goto failed
				}
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
	putBufioWriter(wp, wr)
	if err = conn.Close(); err != nil {
		log.Error("conn.Close(\"%s\", \"%s\") error(%v)", lAddr, rAddr, err)
	}
	// deltimer
	/*
		if _, err = timer.Remove(timerd); err != nil {
			log.Error("\"%s\" dispatch timer.Remove() error(%v)", rAddr, err)
		}
	*/
	log.Debug("\"%s\" dispatch goroutine exit", rAddr)
	return
}

// handshake for goim handshake with client, use rsa & aes.
func (server *Server) handshake(rd *bufio.Reader, wr *bufio.Writer) (aesKey []byte, subKey string, heartbeat time.Duration, bucket *Bucket, channel *Channel, err error) {
	var (
		// TODO
		proto = new(Proto)
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
	if len(aesKey) < aesKeyLen {
		log.Warn("handshake aes key size not valid: %d", len(aesKey))
		err = ErrHandshake
		return
	}
	log.Debug("send handshake response protocol")
	proto.Body = nil
	proto.Operation = OP_HANDSHAKE_REPLY
	if err = server.sendResponse(wr, proto); err != nil {
		log.Error("[%s] server.SendResponse() error(%v)", subKey, err)
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
	if proto.Body, err = aes.ECBDecrypt(proto.Body, aesKey, padding.PKCS7); err != nil {
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

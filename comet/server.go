package main

import (
	"bufio"
	log "code.google.com/p/log4go"
	"errors"
	"github.com/Terry-Mao/goim/libs/crypto/aes"
	"github.com/Terry-Mao/goim/libs/crypto/padding"
	"github.com/Terry-Mao/goim/libs/crypto/rsa"
	"net"
	"sync"
	"time"
)

var (
	defaultOperator = new(IMOperator)
	ErrHandshake    = errors.New("handshake failed")
	aesKeyLen       = 16
	zeroTime        = time.Time{}
	maxInt          = 2 ^ 31 - 1
)

// Request is a header written before every goim connect.  It is used internally
// but documented here as an aid to debugging, such as when analyzing
// network traffic.
type Proto struct {
	PackLen   uint32 // package length
	HeaderLen uint16 // header length
	Ver       uint16 // protocol version
	Operation uint32 // operation for request
	SeqId     uint32 // sequence number chosen by client
	Body      []byte // body
}

type ServerCodec interface {
	ReadRequestHeader(*bufio.Reader, *Proto) error
	ReadRequestBody(*bufio.Reader, *Proto) error
	// WriteResponse must be safe for concurrent use by multiple goroutines.
	WriteResponse(*bufio.Writer, *Proto) error
}

type Operator interface {
	Operate(*Proto) error
	Connect(body []byte) (string, error)
	Disconnect(string) error
}

type Server struct {
	buckets []*Bucket
	rPool   []*sync.Pool
	wPool   []*sync.Pool
	codec   ServerCodec
}

// NewServer returns a new Server.
func NewServer() *Server {
	s := new(Server)
	s.buckets = make([]*Bucket, Conf.Bucket)
	for i := 0; i < Conf.Bucket; i++ {
		s.buckets[i] = NewBucket(Conf.Channel, Conf.CliProto, Conf.SvrProto)
	}
	s.rPool = make([]*sync.Pool, Conf.ReadBuf)
	for i := 0; i < Conf.ReadBuf; i++ {
		s.rPool[i] = new(sync.Pool)
	}
	s.wPool = make([]*sync.Pool, Conf.WriteBuf)
	for i := 0; i < Conf.WriteBuf; i++ {
		s.wPool[i] = new(sync.Pool)
	}
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
		if conn, err = lis.Accept(); err != nil {
			// if listener close then return
			log.Error("listener.Accept() error(%v)", err)
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
		aesKey  []byte // aes key
		subKey  string
		err     error
		bucket  *Bucket
		channel *Channel
		proto   *Proto
		rp      = server.rPool[r&(Conf.ReadBuf-1)]
		wp      = server.wPool[r&(Conf.WriteBuf-1)]
		rd      = newBufioReaderSize(rp, conn, Conf.ReadBuf)
		wr      = newBufioWriterSize(wp, conn, Conf.WriteBuf)
	)
	// handshake
	if aesKey, subKey, bucket, channel, err = server.handshake(conn, rd, wr); err != nil {
		log.Error("handshake() error(%v)", err)
		// may call twice
		if err = conn.Close(); err != nil {
			log.Error("conn.Close() error(%v)", err)
		}
		return
	} else {
		log.Debug("aes key: %v, sub key: \"%s\"", aesKey, subKey)
		// start dispatch goroutine
		go server.dispatch(conn, wr, wp, channel, aesKey)
		for {
			// fetch a proto from channel free list
			if proto, err = channel.CliProto.Set(); err != nil {
				break
			}
			// parse request protocol
			if err = server.readRequest(rd, proto); err != nil {
				break
			}
			// aes decrypt body
			if proto.Body != nil {
				if proto.Body, err = aes.ECBDecrypt(proto.Body, aesKey, padding.PKCS5); err != nil {
					break
				}
			}
			// send to writer
			channel.CliProto.SetAdv()
			select {
			case channel.Signal <- 1:
			default:
				log.Debug("send a signal that channel has message, ignore this time")
				break
			}
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
		log.Error("operator.Operate() error(%v)", err)
	}
	// may call twice
	if err = conn.Close(); err != nil {
		log.Error("conn.Close() error(%v)", err)
	}
	// don't use close chan, Signal can be reused
	// if chan full, writer goroutine next fetch from chan will exit
	// if chan empty, send a 0(close) let the writer exit
	select {
	case channel.Signal <- 0:
	default:
	}
}

// dispatch accepts connections on the listener and serves requests
// for each incoming connection.  dispatch blocks; the caller typically
// invokes it in a go statement.
func (server *Server) dispatch(conn net.Conn, wr *bufio.Writer, wp *sync.Pool, channel *Channel, aesKey []byte) {
	var (
		err    error
		proto  *Proto
		signal int
	)
	log.Debug("start dispatch goroutine")
	for {
		if signal = <-channel.Signal; signal == 0 {
			goto failed
		}
		// fetch message from clibox(client send)
		for {
			if proto, err = channel.CliProto.Get(); err != nil {
				log.Debug("channel no more client message, wait signal")
				break
			}
			// process message
			if err = defaultOperator.Operate(proto); err != nil {
				log.Error("operator.Operate() error(%v)", err)
				goto failed
			}
			if proto.Body != nil {
				if proto.Body, err = aes.ECBEncrypt(proto.Body, aesKey, padding.PKCS5); err != nil {
					goto failed
				}
			}
			if err = server.sendResponse(conn, wr, proto); err != nil {
				log.Error("server.SendResponse() error(%v)", err)
				goto failed
			}
			channel.CliProto.GetAdv()
		}
		// fetch message from svrbox(server send)
		for {
			if proto, err = channel.SvrProto.Get(); err != nil {
				log.Debug("channel no more server message, wait signal")
				break
			}
			if proto.Body != nil {
				if proto.Body, err = aes.ECBEncrypt(proto.Body, aesKey, padding.PKCS5); err != nil {
					goto failed
				}
			}
			// just forward the message
			if err = server.sendResponse(conn, wr, proto); err != nil {
				log.Error("server.SendResponse() error(%v)", err)
				goto failed
			}
			channel.SvrProto.GetAdv()
		}
	}
failed:
	// wake reader up
	putBufioWriter(wp, wr)
	if err = conn.Close(); err != nil {
		log.Error("conn.Close() error(%v)", err)
	}
	log.Debug("dispatch goroutine exit")
	return
}

// handshake for goim handshake with client, use rsa & aes.
func (server *Server) handshake(conn net.Conn, rd *bufio.Reader, wr *bufio.Writer) (aesKey []byte, subKey string, bucket *Bucket, channel *Channel, err error) {
	var (
		body  []byte
		proto Proto
	)
	if err = conn.SetReadDeadline(time.Now().Add(Conf.HandshakeTimeout)); err != nil {
		log.Error("conn.SetReadDeadline() error(%v)", err)
		return
	}
	if err = server.readRequest(rd, &proto); err != nil {
		return
	}
	// a zero value for t means Read will not time out.
	if err = conn.SetReadDeadline(zeroTime); err != nil {
		log.Error("conn.SetReadDeadline() error(%v)", err)
		return
	}
	// TODO rsa decrypt reuse buf?
	log.Debug("cipher body : %v", proto.Body)
	body, err = rsa.Decrypt(proto.Body, RSAPri)
	if err != nil {
		log.Error("Decrypt() error(%v)", err)
		return
	}
	log.Debug("body : %v", body)
	if len(body) < aesKeyLen {
		log.Warn("handshake body size less than %d: %d", aesKeyLen, len(body))
		err = ErrHandshake
		return
	}
	// get aes key use first 16bytes
	aesKey = body[:16]
	// register router
	if subKey, err = defaultOperator.Connect(body[16:]); err != nil {
		log.Error("operator.Operate() error(%v)", err)
		return
	}
	log.Debug("subKey: \"%s\"", subKey)
	proto.Body = nil
	// update subkey -> channel
	bucket = server.bucket(subKey)
	// channel = bucket.GetChannel()
	// channel.Reset()
	// TODO how to reuse channel
	channel = NewChannel(Conf.CliProto, Conf.SvrProto)
	bucket.Put(subKey, channel)
	proto.Operation = OP_HANDSHARE_REPLY
	if err = server.sendResponse(conn, wr, &proto); err != nil {
		log.Error("server.SendResponse() error(%v)", err)
	}
	return
}

// readRequest
func (server *Server) readRequest(rd *bufio.Reader, proto *Proto) (err error) {
	log.Debug("readRequestHeader")
	if err = server.readRequestHeader(rd, proto); err != nil {
		return
	}
	// read body
	log.Debug("readRequestBody")
	if err = server.codec.ReadRequestBody(rd, proto); err != nil {
	}
	log.Debug("proto: %v", proto)
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
func (server *Server) sendResponse(conn net.Conn, wr *bufio.Writer, proto *Proto) (err error) {
	if err = conn.SetWriteDeadline(time.Now().Add(Conf.WriteTimeout)); err != nil {
		log.Error("conn.SetWriteDeadline() error(%v)", err)
		return
	}
	if err = server.codec.WriteResponse(wr, proto); err != nil {
		log.Error("server.codec.WriteResponse() error(%v)", err)
	}
	return
}

func (server *Server) bucket(subKey string) *Bucket {
	h := NewMurmur3C()
	h.Write([]byte(subKey))
	idx := h.Sum32() & uint32(Conf.Bucket-1)
	log.Debug("sub key:\"%s\" hit channel bucket index:%d", subKey, idx)
	return server.buckets[idx]
}

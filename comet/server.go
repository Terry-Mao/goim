package main

import (
	"bufio"
	log "code.google.com/p/log4go"
	"errors"
	"net"
	"sync"
)

var (
	defaultOperator = new(IMOperator)
	ErrHandshake    = errors.New("handshake failed")
	aesKeyLen       = 16
)

// Request is a header written before every goim connect.  It is used internally
// but documented here as an aid to debugging, such as when analyzing
// network traffic.
type Proto struct {
	Ver       uint16 // protocol version
	Operation uint32 // operation for request
	SeqId     uint32 // sequence number chosen by client
	Body      []byte // body
	next      *Proto // for free list in Server
}

type ServerCodec interface {
	ReadRequestHeader(*Proto) error
	ReadRequestBody() ([]byte, error)
	// WriteResponse must be safe for concurrent use by multiple goroutines.
	WriteResponse(*Proto) error

	Close() error
}

type Operator interface {
	Operate(*Proto) error
	Connect(body []byte) (string, error)
	Disconnect(string) error
}

type Server struct {
	pLock sync.Mutex // protects freeProto
	free  *Proto
}

// NewServer returns a new Server.
func NewServer() *Server {
	return new(Server)
}

// Accept accepts connections on the listener and serves requests
// for each incoming connection.  Accept blocks; the caller typically
// invokes it in a go statement.
func (server *Server) Accept(lis net.Listener) {
	var (
		conn net.Conn
		err  error
	)
	for {
		if conn, err = lis.Accept(); err != nil {
			// if listener close then return
			log.Error("listener.Accept() error(%v)", err)
			return
		}
		go server.ServeConn(conn)
	}
}

// ServeConn runs the server on a single connection.
// ServeConn blocks, serving the connection until the client hangs up.
// The caller typically invokes ServeConn in a go statement.
// ServeConn uses the goim wire format on the
// connection.  To use an alternate codec, use ServeCodec.
func (server *Server) ServeConn(conn net.Conn) {
	// TODO reuse buf
	srv := &IMServerCodec{
		conn:  conn,
		rdBuf: bufio.NewReader(conn),
		wrBuf: bufio.NewWriter(conn),
	}
	server.ServeCodec(srv)
}

// ServeCodec is like ServeConn but uses the specified codec to
// decode requests and encode responses.
func (server *Server) ServeCodec(codec ServerCodec) {
	var (
		aesKey []byte // aes key
		subKey string
		err    error
		msg    = make(chan *Proto, 1024)
		proto  *Proto
		closed = false
	)
	// handshake
	if aesKey, subKey, err = server.handshake(codec); err != nil {
		log.Error("handshake() error(%v)", err)
	} else {
		log.Debug("aes key: %v, sub key: \"%s\"", aesKey, subKey)
		// start dispatch goroutine
		go server.dispatch(msg, &closed, codec)
		for {
			proto, err = server.readRequest(codec)
			if err != nil {
				break
			}
			// decrypt body
			msg <- proto
		}
	}
	if !closed {
		closed = true
		// wake writer
		close(msg)
		if err = codec.Close(); err != nil {
			log.Error("codec.Close() error(%v)", err)
		}
		log.Info("close")
	}
	// disconnect, revoke router
	if err = defaultOperator.Disconnect(subKey); err != nil {
		log.Error("operator.Operate() error(%v)", err)
	}
}

// dispatch accepts connections on the listener and serves requests
// for each incoming connection.  dispatch blocks; the caller typically
// invokes it in a go statement.
func (server *Server) dispatch(msg chan *Proto, closed *bool, codec ServerCodec) {
	var (
		err   error
		proto *Proto
	)
	log.Debug("start dispatch goroutine")
	for {
		proto = <-msg
		if proto == nil {
			// woken by reader
			return
		}
		// process message
		if err = defaultOperator.Operate(proto); err != nil {
			log.Error("operator.Operate() error(%v)", err)
			goto failed
		}
		if err = server.sendResponse(proto, codec); err != nil {
			log.Error("server.SendResponse() error(%v)", err)
			goto failed
		}
		continue
	failed:
		// writer close, reader skip
		*closed = true
		if err = codec.Close(); err != nil {
			log.Error("codec.Close() error(%v)", err)
		}
		log.Debug("dispatch goroutine exit")
		return
	}
}

// handshake for goim handshake with client, use rsa & aes.
func (server *Server) handshake(codec ServerCodec) (aeskey []byte, subKey string, err error) {
	var (
		body  []byte
		proto *Proto
	)
	proto, err = server.readRequest(codec)
	if err != nil {
		return
	}
	// rsa decrypt TODO reuse buf?
	body, err = Decrypt(proto.Body)
	if err != nil {
		log.Error("Decrypt() error(%v)", err)
		return
	}
	// get aes key use first 16bytes
	if len(body) < aesKeyLen {
		log.Warn("handshake body size less than %d: %d", aesKeyLen, len(body))
		err = ErrHandshake
		return
	}
	// register router
	if subKey, err = defaultOperator.Connect(body[16:]); err != nil {
		log.Error("operator.Operate() error(%v)", err)
		return
	}
	log.Debug("subKey: \"%s\"", subKey)
	// TODO update map
	// reply client
	proto.Body = nil
	proto.Operation = OP_HANDSHARE_REPLY
	if err = server.sendResponse(proto, codec); err != nil {
		log.Error("server.SendResponse() error(%v)", err)
		return
	}
	return body[:16], subKey, nil
}

// readRequest
func (server *Server) readRequest(codec ServerCodec) (proto *Proto, err error) {
	log.Debug("readRequestHeader")
	if proto, err = server.readRequestHeader(codec); err != nil {
		return
	}
	// read body
	log.Debug("readRequestBody")
	if proto.Body, err = codec.ReadRequestBody(); err != nil {
	}
	return
}

func (server *Server) readRequestHeader(codec ServerCodec) (proto *Proto, err error) {
	// Grab the request header.
	proto = server.getProto()
	if err = codec.ReadRequestHeader(proto); err != nil {
		proto = nil
	}
	return
}

// sendResponse send resp to client, sendResponse must be goroutine safe.
func (server *Server) sendResponse(proto *Proto, codec ServerCodec) (err error) {
	err = codec.WriteResponse(proto)
	server.freeProto(proto)
	return
}

func (server *Server) getProto() *Proto {
	server.pLock.Lock()
	proto := server.free
	if proto == nil {
		proto = new(Proto)
	} else {
		server.free = proto.next
		*proto = Proto{} // reset
	}
	server.pLock.Unlock()
	return proto
}

func (server *Server) freeProto(proto *Proto) {
	server.pLock.Lock()
	proto.next = server.free
	server.free = proto
	server.pLock.Unlock()
}

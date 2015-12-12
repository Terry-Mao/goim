package main

import (
	"bufio"
	log "code.google.com/p/log4go"
	"encoding/json"
	"github.com/Terry-Mao/goim/libs/define"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"time"
)

func InitHTTP(addrs []string) (err error) {
	var (
		bind         string
		listener     *net.TCPListener
		addr         *net.TCPAddr
		server       *http.Server
		httpServeMux = http.NewServeMux()
	)
	httpServeMux.HandleFunc("/sub", serveHTTP)
	for _, bind = range addrs {
		if addr, err = net.ResolveTCPAddr("tcp4", bind); err != nil {
			log.Error("net.ResolveTCPAddr(\"tcp4\", \"%s\") error(%v)", bind, err)
			return
		}
		if listener, err = net.ListenTCP("tcp4", addr); err != nil {
			log.Error("net.ListenTCP(\"tcp4\", \"%s\") error(%v)", bind, err)
			return
		}
		server = &http.Server{Handler: httpServeMux}
		if Debug {
			log.Debug("start http listen: \"%s\"", bind)
		}
		go func() {
			if err = server.Serve(listener); err != nil {
				log.Error("server.Serve(\"%s\") error(%v)", bind, err)
				panic(err)
			}
		}()
	}
	return
}

func serveHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		// ip addr
		rAddr = r.RemoteAddr
		// timer
		tr = DefaultServer.round.Timer(rand.Int())
	)
	log.Debug("start websocket serve with \"%s\"", rAddr)
	DefaultServer.serveHTTP(w, r, tr)
}

func (server *Server) serveHTTP(w http.ResponseWriter, r *http.Request, tr *Timer) {
	var (
		b    *Bucket
		ok   bool
		hb   time.Duration // heartbeat
		key  string
		cb   string
		err  error
		trd  *TimerData
		conn net.Conn
		rwr  *bufio.ReadWriter
		hj   http.Hijacker
		// no client send
		ch = NewChannel(1, define.NoRoom)
	)
	if key, cb, hb, err = server.authHTTP(r, ch); err != nil {
		http.Error(w, "auth failed", http.StatusForbidden)
		return
	}
	if hj, ok = w.(http.Hijacker); !ok {
		log.Error("w.(http.Hijacker) type assection failed")
		http.Error(w, "not support", http.StatusInternalServerError)
		return
	}
	if conn, rwr, err = hj.Hijack(); err != nil {
		log.Error("hj.Hijack() error(%v)", err)
		http.Error(w, "not support", http.StatusInternalServerError)
		return
	}
	// register key->channel
	b = server.Bucket(key)
	b.Put(key, ch)
	// hanshake ok start dispatch goroutine
	server.dispatchHTTP(rwr, cb, ch, hb)
	tr.Del(trd)
	conn.Close()
	b.Del(key)
	// don't use close chan, Signal can be reused
	// if chan full, writer goroutine next fetch from chan will exit
	// if chan empty, send a 0(close) let the writer exit
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
func (server *Server) dispatchHTTP(rwr *bufio.ReadWriter, cb string, ch *Channel, hb time.Duration) {
	var (
		p   *Proto
		err error
	)
	if Debug {
		log.Debug("start dispatch goroutine")
	}
	if !ch.ReadyWithTimeout(hb) {
		return
	}
	// fetch message from svrbox(server send)
	if p, err = ch.SvrProto.Get(); err != nil {
		return
	}
	// just forward the message
	if err = server.writeHTTPResponse(rwr, cb, p); err != nil {
		return
	}
	ch.SvrProto.GetAdv()
	return
}

// auth for goim handshake with client, use rsa & aes.
func (server *Server) authHTTP(r *http.Request, ch *Channel) (subKey, callback string, heartbeat time.Duration, err error) {
	var (
		pStr   string
		pInt   int64
		params = r.URL.Query()
		p      = &ch.CliProto
	)
	pStr = params.Get("ver")
	if pInt, err = strconv.ParseInt(pStr, 10, 16); err != nil {
		log.Error("strconv.ParseInt(\"%s\", 10) error(%v)", err)
		return
	}
	p.Ver = int16(pInt)
	pStr = params.Get("op")
	if pInt, err = strconv.ParseInt(pStr, 10, 32); err != nil {
		log.Error("strconv.ParseInt(\"%s\", 10) error(%v)", err)
		return
	}
	p.Operation = int32(pInt)
	pStr = params.Get("seq")
	if pInt, err = strconv.ParseInt(pStr, 10, 32); err != nil {
		log.Error("strconv.ParseInt(\"%s\", 10) error(%v)", err)
		return
	}
	p.SeqId = int32(pInt)
	if p.Operation != define.OP_AUTH {
		log.Warn("auth operation not valid: %d", p.Operation)
		err = ErrOperation
		return
	}
	callback = params.Get("cb")
	p.Body = []byte(params.Get("t"))
	if subKey, ch.RoomId, heartbeat, err = server.operator.Connect(p); err != nil {
		log.Error("operator.Connect error(%v)", err)
	}
	return
}

// sendResponse send resp to client, sendResponse must be goroutine safe.
func (server *Server) writeHTTPResponse(rwr *bufio.ReadWriter, cb string, proto *Proto) (err error) {
	var pb []byte
	if proto.Body == nil {
		proto.Body = emptyJSONBody
	}
	if pb, err = json.Marshal(proto); err != nil {
		log.Error("json.Marshal() error(%v)", err)
		return
	}
	if len(cb) != 0 {
		if _, err = rwr.WriteString(cb); err != nil {
			return
		}
		if err = rwr.WriteByte('='); err != nil {
			return
		}
	}
	if _, err = rwr.Write(pb); err != nil {
		return
	}
	err = rwr.Flush()
	return
}

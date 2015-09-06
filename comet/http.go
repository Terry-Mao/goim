package main

import (
	"bufio"
	log "code.google.com/p/log4go"
	"encoding/json"
	"github.com/Terry-Mao/goim/define"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"time"
)

const (
	maxPackLen    = 1 << 10
	rawHeaderLen  = int16(16)
	packLenSize   = 4
	headerLenSize = 2
)

func InitHTTP() (err error) {
	var (
		listener     *net.TCPListener
		addr         *net.TCPAddr
		httpServeMux = http.NewServeMux()
	)
	httpServeMux.HandleFunc("/sub", serveHTTP)
	for _, bind := range Conf.HTTPBind {
		if addr, err = net.ResolveTCPAddr("tcp4", bind); err != nil {
			log.Error("net.ResolveTCPAddr(\"tcp4\", \"%s\") error(%v)", bind, err)
			return
		}
		if listener, err = net.ListenTCP("tcp4", addr); err != nil {
			log.Error("net.ListenTCP(\"tcp4\", \"%s\") error(%v)", bind, err)
			return
		}
		server := &http.Server{Handler: httpServeMux}
		log.Debug("start http listen: \"%s\"", bind)
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
		ch = NewChannel(0, 1, define.NoRoom)
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
	if trd, err = tr.Add(hb, conn); err != nil {
		log.Error("handshake: timer.Add() error(%v)", err)
		if err = conn.Close(); err != nil {
			log.Error("handshake: conn.Close() error(%v)", err)
		}
		return
	}
	// register key->channel
	b = server.Bucket(key)
	b.Put(key, ch)
	// hanshake ok start dispatch goroutine
	server.dispatchHTTP(rwr, cb, ch)
	// dialog finish
	// revoke the subkey
	// revoke the remote subkey
	// close the net.Conn
	// read & write goroutine
	// return channel to bucket's free list
	if err = conn.Close(); err != nil {
		log.Error("conn.Close() error(%v)", err)
	}
	b.Del(key)
	tr.Del(trd)
	// don't use close chan, Signal can be reused
	// if chan full, writer goroutine next fetch from chan will exit
	// if chan empty, send a 0(close) let the writer exit
	if err = server.operator.Disconnect(key, ch.RoomId); err != nil {
		log.Error("%s operator do disconnect error(%v)", key, err)
	}
	log.Debug("%s serverconn goroutine exit", key)
	return
}

// dispatch accepts connections on the listener and serves requests
// for each incoming connection.  dispatch blocks; the caller typically
// invokes it in a go statement.
func (server *Server) dispatchHTTP(rwr *bufio.ReadWriter, cb string, ch *Channel) {
	var (
		p   *Proto
		err error
	)
	log.Debug("start dispatch goroutine")
	if !ch.Ready() {
		return
	}
	// fetch message from svrbox(server send)
	if p, err = ch.SvrProto.Get(); err != nil {
		log.Debug("channel no more server message, wait signal")
		return
	}
	// just forward the message
	if err = server.writeHTTPResponse(rwr, cb, p); err != nil {
		log.Error("server.sendTCPResponse() error(%v)", err)
		return
	}
	ch.SvrProto.GetAdv()
	return
}

// auth for goim handshake with client, use rsa & aes.
func (server *Server) authHTTP(r *http.Request, ch *Channel) (subKey, callback string, heartbeat time.Duration, err error) {
	var (
		p      *Proto
		pStr   string
		pInt   int64
		params = r.URL.Query()
	)
	// WARN
	// don't adv the svr(no client send) proto, after auth simply discard it.
	if p, err = ch.SvrProto.Set(); err != nil {
		return
	}
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
			log.Error("http rwr.Write() error(%v)", err)
			return
		}
		if err = rwr.WriteByte('='); err != nil {
			log.Error("http rwr.Write() error(%v)", err)
			return
		}
	}
	if _, err = rwr.Write(pb); err != nil {
		log.Error("http rwr.Write() error(%v)", err)
		return
	}
	if err = rwr.Flush(); err != nil {
		log.Error("http rwr.Flush() error(%v)", err)
	}
	proto.Reset()
	return
}

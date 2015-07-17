package main

/*
import (
	"bufio"
	log "code.google.com/p/log4go"
	"net"
	"net/http"
	"time"
)

const (
	maxPackLen    = 1 << 10
	rawHeaderLen  = int16(16)
	packLenSize   = 4
	headerLenSize = 2
)

func (server *Server) serveHttp(w http.ResponseWriter, r *http.Request, tr *Timer) {
	var (
		b    *Bucket
		ch   *Channel
		hb   time.Duration // heartbeat
		key  string
		err  error
		trd  *TimerData
		conn *net.TCPConn
		rwr  *bufio.ReadWriter
		hj   http.Hijacker
		ok   bool
	)
	if hj, ok = w.(http.Hijacker); !ok {
		http.Error(w, "webserver doesn't support hijacking", http.StatusInternalServerError)
		return
	}
	if conn, rwr, err = hj.Hijack(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// auth
	if trd, err = tr.Add(Conf.HandshakeTimeout, conn); err != nil {
		log.Error("handshake: timer.Add() error(%v)", err)
	} else {
		if key, hb, err = server.authHTTP(r, p); err != nil {
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
		return
	}
	// TODO how to reuse channel
	// register key->channel
	b = server.Bucket(key)
	// no client send
	ch = NewChannel(0, Conf.SvrProto)
	b.Put(key, ch)
	// hanshake ok start dispatch goroutine
	server.dispatchHTTP(conn, wr, wpb, ch, hb, tr)
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
	// don't use close chan, Signal can be reused
	// if chan full, writer goroutine next fetch from chan will exit
	// if chan empty, send a 0(close) let the writer exit
	if err = server.operator.Disconnect(key); err != nil {
		log.Error("%s operator do disconnect error(%v)", key, err)
	}
	log.Debug("%s serverconn goroutine exit", key)
	return
}

// dispatch accepts connections on the listener and serves requests
// for each incoming connection.  dispatch blocks; the caller typically
// invokes it in a go statement.
func (server *Server) dispatchHTTP(conn *net.TCPConn, rwr *bufio.ReadWriter, ch *Channel, hb time.Duration, tr *Timer) {
	var (
		p   *Proto
		err error
		trd *TimerData
		sig int
	)
	log.Debug("start dispatch goroutine")
	// fetch message from svrbox(server send)
	for {
		if p, err = ch.SvrProto.Get(); err != nil {
			log.Debug("channel no more server message, wait signal")
			break
		}
		// just forward the message
		if err = server.writeTCPResponse(rwr, p); err != nil {
			log.Error("server.sendTCPResponse() error(%v)", err)
			goto failed
		}
		ch.SvrProto.GetAdv()
	}
failed:
	// deltimer
	tr.Del(trd)
	log.Debug("dispatch goroutine exit")
	return
}

// auth for goim handshake with client, use rsa & aes.
func (server *Server) authHTTP(rr *bufio.Reader, wr *bufio.Writer, pb []byte, p *Proto) (subKey string, heartbeat time.Duration, err error) {
	log.Debug("get auth request protocol")
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
	log.Debug("send auth response protocol")
	p.Body = nil
	p.Operation = OP_AUTH_REPLY
	if err = server.writeTCPResponse(wr, pb, p); err != nil {
		log.Error("[%s] server.sendTCPResponse() error(%v)", subKey, err)
	}
	return
}

// sendResponse send resp to client, sendResponse must be goroutine safe.
func (server *Server) writeHTTPResponse(rwr *bufio.Writer, proto *Proto) (err error) {
	return
}
*/

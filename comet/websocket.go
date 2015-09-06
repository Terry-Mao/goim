package main

import (
	log "code.google.com/p/log4go"
	"github.com/Terry-Mao/goim/define"
	"golang.org/x/net/websocket"
	"math/rand"
	"net"
	"net/http"
	"time"
)

func InitWebsocket() (err error) {
	var (
		listener     *net.TCPListener
		addr         *net.TCPAddr
		httpServeMux = http.NewServeMux()
	)
	httpServeMux.Handle("/sub", websocket.Handler(serveWebsocket))
	for _, bind := range Conf.WebsocketBind {
		if addr, err = net.ResolveTCPAddr("tcp4", bind); err != nil {
			log.Error("net.ResolveTCPAddr(\"tcp4\", \"%s\") error(%v)", bind, err)
			return
		}
		if listener, err = net.ListenTCP("tcp4", addr); err != nil {
			log.Error("net.ListenTCP(\"tcp4\", \"%s\") error(%v)", bind, err)
			return
		}
		server := &http.Server{Handler: httpServeMux}
		log.Debug("start websocket listen: \"%s\"", bind)
		go func() {
			if err = server.Serve(listener); err != nil {
				log.Error("server.Serve(\"%s\") error(%v)", bind, err)
				panic(err)
			}
		}()
	}
	return
}

func serveWebsocket(conn *websocket.Conn) {
	var (
		// ip addr
		lAddr = conn.LocalAddr()
		rAddr = conn.RemoteAddr()
		// timer
		tr = DefaultServer.round.Timer(rand.Int())
	)
	log.Debug("start websocket serve \"%s\" with \"%s\"", lAddr, rAddr)
	DefaultServer.serveWebsocket(conn, tr)
}

func (server *Server) serveWebsocket(conn *websocket.Conn, tr *Timer) {
	var (
		p   *Proto
		b   *Bucket
		hb  time.Duration // heartbeat
		key string
		err error
		trd *TimerData
		ch  = NewChannel(Conf.CliProto, Conf.SvrProto, define.NoRoom)
	)
	// auth
	if trd, err = tr.Add(Conf.HandshakeTimeout, conn); err != nil {
		log.Error("handshake: timer.Add() error(%v)", err)
	} else {
		if key, hb, err = server.authWebsocket(conn, ch); err != nil {
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
	// register key->channel
	b = server.Bucket(key)
	b.Put(key, ch)
	// hanshake ok start dispatch goroutine
	go server.dispatchWebsocket(conn, ch, hb, tr)
	for {
		// fetch a proto from channel free list
		if p, err = ch.CliProto.Set(); err != nil {
			log.Error("%s fetch client proto error(%v)", key, err)
			break
		}
		// parse request protocol
		if err = server.readWebsocketRequest(conn, p); err != nil {
			log.Error("%s read client request error(%v)", key, err)
			break
		}
		// send to writer
		ch.CliProto.SetAdv()
		ch.Signal()
	}
	// dialog finish
	// revoke the subkey
	// revoke the remote subkey
	// close the net.Conn
	// read & write goroutine
	// return channel to bucket's free list
	// may call twice
	if err = conn.Close(); err != nil {
		log.Error("reader: conn.Close() error(%v)")
	}
	ch.Finish()
	b.Del(key)
	if err = server.operator.Disconnect(key, ch.RoomId); err != nil {
		log.Error("%s operator do disconnect error(%v)", key, err)
	}
	log.Debug("%s serverconn goroutine exit", key)
	return
}

// dispatch accepts connections on the listener and serves requests
// for each incoming connection.  dispatch blocks; the caller typically
// invokes it in a go statement.
func (server *Server) dispatchWebsocket(conn *websocket.Conn, ch *Channel, hb time.Duration, tr *Timer) {
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
			if err = server.writeWebsocketResponse(conn, p); err != nil {
				log.Error("server.sendTCPResponse() error(%v)", err)
				goto failed
			}
			ch.CliProto.GetAdv()
		}
		// fetch message from svrbox(server send)
		for {
			if p, err = ch.SvrProto.Get(); err != nil {
				log.Warn("ch.SvrProto.Get() error(%v)", err)
				break
			}
			// just forward the message
			if err = server.writeWebsocketResponse(conn, p); err != nil {
				log.Error("server.sendTCPResponse() error(%v)", err)
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
	log.Debug("dispatch goroutine exit")
	return
}

// auth for goim handshake with client, use rsa & aes.
func (server *Server) authWebsocket(conn *websocket.Conn, ch *Channel) (subKey string, heartbeat time.Duration, err error) {
	var p *Proto
	// WARN
	// don't adv the cli proto, after auth simply discard it.
	if p, err = ch.CliProto.Set(); err != nil {
		return
	}
	if err = server.readWebsocketRequest(conn, p); err != nil {
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
	if err = server.writeWebsocketResponse(conn, p); err != nil {
		log.Error("[%s] server.sendTCPResponse() error(%v)", subKey, err)
	}
	return
}

// readRequest
func (server *Server) readWebsocketRequest(conn *websocket.Conn, proto *Proto) (err error) {
	if err = websocket.JSON.Receive(conn, proto); err != nil {
		log.Error("websocket.JSON.Receive() error(%v)", err)
	}
	return
}

// sendResponse send resp to client, sendResponse must be goroutine safe.
func (server *Server) writeWebsocketResponse(conn *websocket.Conn, proto *Proto) (err error) {
	if proto.Body == nil {
		proto.Body = emptyJSONBody
	}
	if err = websocket.JSON.Send(conn, proto); err != nil {
		log.Error("websocket.JSON.Send() error(%v)", err)
	}
	proto.Reset()
	return
}

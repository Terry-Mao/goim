package main

import (
	log "code.google.com/p/log4go"
	"golang.org/x/net/websocket"
	"time"
)

func (server *Server) serveWebsocket(conn *websocket.Conn, tr *Timer) {
	var (
		b   *Bucket
		ch  *Channel
		hb  time.Duration // heartbeat
		key string
		err error
		trd *TimerData
		p   = new(Proto)
	)
	// auth
	if trd, err = tr.Add(Conf.HandshakeTimeout, conn); err != nil {
		log.Error("handshake: timer.Add() error(%v)", err)
	} else {
		if key, hb, err = server.authWebsocket(conn, p); err != nil {
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
	ch = NewChannel(Conf.CliProto, Conf.SvrProto)
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
	// may call twice
	if err = conn.Close(); err != nil {
		log.Error("reader: conn.Close() error(%v)")
	}
	// don't use close chan, Signal can be reused
	// if chan full, writer goroutine next fetch from chan will exit
	// if chan empty, send a 0(close) let the writer exit
	select {
	case ch.Signal <- ProtoFinsh:
	default:
		log.Warn("%s send proto finish signal, but chan is full just ignore", key)
	}
	b.Del(key)
	if err = server.operator.Disconnect(key); err != nil {
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
		sig int
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
			if err = server.writeWebsocketResponse(conn, p); err != nil {
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
		log.Error("conn.Close() error(%v)", err)
	}
	// deltimer
	tr.Del(trd)
	log.Debug("dispatch goroutine exit")
	return
}

// auth for goim handshake with client, use rsa & aes.
func (server *Server) authWebsocket(conn *websocket.Conn, p *Proto) (subKey string, heartbeat time.Duration, err error) {
	if err = server.readWebsocketRequest(conn, p); err != nil {
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
	return
}

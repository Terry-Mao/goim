package main

import (
	log "code.google.com/p/log4go"
	"crypto/tls"
	"github.com/Terry-Mao/goim/libs/define"
	"golang.org/x/net/websocket"
	"math/rand"
	"net"
	"net/http"
	"time"
)

func InitWebsocket(addrs []string) (err error) {
	var (
		bind         string
		listener     *net.TCPListener
		addr         *net.TCPAddr
		httpServeMux = http.NewServeMux()
		server       *http.Server
	)
	httpServeMux.Handle("/sub", websocket.Handler(serveWebsocket))
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
			log.Debug("start websocket listen: \"%s\"", bind)
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

func InitWebsocketWithTLS(addrs []string, cert, priv string) (err error) {
	var (
		httpServeMux = http.NewServeMux()
	)
	httpServeMux.Handle("/sub", websocket.Handler(serveWebsocket))
	config := &tls.Config{}
	config.Certificates = make([]tls.Certificate, 1)
	config.Certificates[0], err = tls.LoadX509KeyPair(cert, priv)
	if err != nil {
		return
	}
	for _, bind := range addrs {
		server := &http.Server{Addr: bind, Handler: httpServeMux}
		server.SetKeepAlivesEnabled(true)
		if Debug {
			log.Debug("start websocket wss listen: \"%s\"", bind)
		}
		go func() {
			ln, err := net.Listen("tcp", bind)
			if err != nil {
				return
			}

			tlsListener := tls.NewListener(ln, config)
			if err = server.Serve(tlsListener); err != nil {
				log.Error("server.Serve(\"%s\") error(%v)", bind, err)
				return
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
		ch  = NewChannel(server.Options.SvrProto, define.NoRoom)
	)
	// handshake
	if trd, err = tr.Add(server.Options.HandshakeTimeout, conn); err == nil {
		if key, hb, err = server.authWebsocket(conn, ch); err == nil {
			tr.Set(trd, hb)
		}
	}
	if err != nil {
		log.Error("handshake failed error(%v)", err)
		if trd != nil {
			tr.Del(trd)
		}
		conn.Close()
		return
	}
	// register key->channel
	b = server.Bucket(key)
	b.Put(key, ch)
	// hanshake ok start dispatch goroutine
	go server.dispatchWebsocket(conn, ch)
	for {
		// parse request protocol
		if err = server.readWebsocketRequest(conn, p); err != nil {
			break
		}
		if p.Operation == define.OP_HEARTBEAT {
			// Use a previous timer value if difference between it and a new
			// value is less than TIMER_LAZY_DELAY milliseconds: this allows
			// to minimize the minheap operations for fast connections.
			if !trd.Lazy(hb) {
				tr.Set(trd, hb)
			}
			// heartbeat
			p.Body = nil
			p.Operation = define.OP_HEARTBEAT_REPLY
		} else {
			// process message
			if err = server.operator.Operate(p); err != nil {
				break
			}
		}
		if err = server.writeWebsocketResponse(conn, p); err != nil {
			break
		}
	}
	conn.Close()
	ch.Close()
	b.Del(key)
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
func (server *Server) dispatchWebsocket(conn *websocket.Conn, ch *Channel) {
	var (
		p   *Proto
		err error
	)
	if Debug {
		log.Debug("start dispatch goroutine")
	}
	for {
		if !ch.Ready() {
			goto failed
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
	if Debug {
		log.Debug("dispatch goroutine exit")
	}
	return
}

func (server *Server) authWebsocket(conn *websocket.Conn, ch *Channel) (key string, heartbeat time.Duration, err error) {
	var p = &ch.CliProto
	if err = server.readWebsocketRequest(conn, p); err != nil {
		return
	}
	if p.Operation != define.OP_AUTH {
		err = ErrOperation
		return
	}
	if key, ch.RoomId, heartbeat, err = server.operator.Connect(p); err != nil {
		return
	}
	p.Body = nil
	p.Operation = define.OP_AUTH_REPLY
	err = server.writeWebsocketResponse(conn, p)
	return
}

func (server *Server) readWebsocketRequest(conn *websocket.Conn, p *Proto) (err error) {
	p.Reset()
	err = websocket.JSON.Receive(conn, p)
	return
}

func (server *Server) writeWebsocketResponse(conn *websocket.Conn, p *Proto) (err error) {
	if p.Body == nil {
		p.Body = emptyJSONBody
	}
	err = websocket.JSON.Send(conn, p)
	return
}

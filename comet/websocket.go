package main

import (
	"crypto/tls"
	"math/rand"
	"net"
	"net/http"
	"time"

	log "code.google.com/p/log4go"
	"github.com/Terry-Mao/goim/libs/define"
	"github.com/Terry-Mao/goim/libs/proto"
	itime "github.com/Terry-Mao/goim/libs/time"
	"golang.org/x/net/websocket"
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
	if config.Certificates[0], err = tls.LoadX509KeyPair(cert, priv); err != nil {
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

func (server *Server) serveWebsocket(conn *websocket.Conn, tr *itime.Timer) {
	var (
		err error
		key string
		hb  time.Duration // heartbeat
		p   *proto.Proto
		b   *Bucket
		trd *itime.TimerData
		ch  = NewChannel(server.Options.CliProto, server.Options.SvrProto, define.NoRoom)
	)
	// handshake
	trd = tr.Add(server.Options.HandshakeTimeout, func() {
		conn.Close()
	})
	// must not setadv, only used in auth
	if p, err = ch.CliProto.Set(); err == nil {
		key, ch.RoomId, hb, err = server.authWebsocket(conn, p)
	}
	if err != nil {
		conn.Close()
		tr.Del(trd)
		log.Error("handshake failed error(%v)", err)
		return
	}
	trd.Key = key
	tr.Set(trd, hb)
	b = server.Bucket(key)
	b.Put(key, ch, tr)
	// hanshake ok start dispatch goroutine
	go server.dispatchWebsocket(key, conn, ch)
	for {
		if p, err = ch.CliProto.Set(); err != nil {
			break
		}
		if err = p.ReadWebsocket(conn); err != nil {
			break
		}
		if p.Operation == define.OP_HEARTBEAT {
			// heartbeat
			tr.Set(trd, hb)
			p.Body = nil
			p.Operation = define.OP_HEARTBEAT_REPLY
		} else {
			// process message
			if err = server.operator.Operate(p); err != nil {
				break
			}
		}
		ch.CliProto.SetAdv()
		ch.Signal()
	}
	log.Error("key: %s server websocket failed error(%v)", key, err)
	conn.Close()
	ch.Close()
	b.Del(key)
	if err = server.operator.Disconnect(key, ch.RoomId); err != nil {
		log.Error("key: %s operator do disconnect error(%v)", key, err)
	}
	if Debug {
		log.Debug("key: %s server websocket goroutine exit", key)
	}
	return
}

// dispatch accepts connections on the listener and serves requests
// for each incoming connection.  dispatch blocks; the caller typically
// invokes it in a go statement.
func (server *Server) dispatchWebsocket(key string, conn *websocket.Conn, ch *Channel) {
	var (
		p   *proto.Proto
		err error
	)
	if Debug {
		log.Debug("key: %s start dispatch websocket goroutine", key)
	}
	for {
		if !ch.Ready() {
			if Debug {
				log.Debug("key: %s wakeup exit dispatch goroutine", key)
			}
			break
		}
		// fetch message from svrbox(server send)
		for {
			if p, err = ch.SvrProto.Get(); err != nil {
				log.Warn("ch.SvrProto.Get() error(%v)", err)
				err = nil
				break
			}
			// just forward the message
			if err = p.WriteWebsocket(conn); err != nil {
				log.Error("server.sendTCPResponse() error(%v)", err)
				goto failed
			}
			p.Body = nil // avoid memory leak
			ch.SvrProto.GetAdv()
		}
	}
failed:
	log.Error("key: %s dispatch websocket error(%v)", key, err)
	if err = conn.Close(); err != nil {
		log.Warn("conn.Close() error(%v)", err)
	}
	if Debug {
		log.Debug("key: %s dispatch goroutine exit", key)
	}
	return
}

func (server *Server) authWebsocket(conn *websocket.Conn, p *proto.Proto) (key string, rid int32, heartbeat time.Duration, err error) {
	if err = p.ReadWebsocket(conn); err != nil {
		return
	}
	if p.Operation != define.OP_AUTH {
		err = ErrOperation
		return
	}
	if key, rid, heartbeat, err = server.operator.Connect(p); err != nil {
		return
	}
	p.Body = nil
	p.Operation = define.OP_AUTH_REPLY
	err = p.WriteWebsocket(conn)
	return
}

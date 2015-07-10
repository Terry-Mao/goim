package main

import (
	log "code.google.com/p/log4go"
	"golang.org/x/net/websocket"
	"math/rand"
	"net"
	"net/http"
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
		rAddr = conn.Request().RemoteAddr
	)
	log.Debug("start serve \"%s\"", rAddr)
	DefaultServer.serve(conn, conn, nil, conn, rand.Int())
}

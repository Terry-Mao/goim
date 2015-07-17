package main

import (
	log "code.google.com/p/log4go"
	"math/rand"
	"net"
	"net/http"
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

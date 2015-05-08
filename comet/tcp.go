package main

import (
	log "code.google.com/p/log4go"
	"net"
)

// InitTCP listen all tcp.bind and start accept connections.
func InitTCP(server *Server) (err error) {
	var (
		listener net.Listener
	)
	for _, bind := range Conf.TCPBind {
		if listener, err = net.Listen("tcp", bind); err != nil {
			log.Error("net.ListenTCP(\"tcp4\", \"%s\") error(%v)", bind, err)
			return
		}
		// split N core accept
		for i := 0; i < Conf.MaxProc; i++ {
			log.Debug("start tcp accept[goroutine %d]: \"%s\"", i, bind)
			go server.Accept(listener)
		}
	}
	return
}

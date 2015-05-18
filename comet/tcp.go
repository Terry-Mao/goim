package main

import (
	log "code.google.com/p/log4go"
	"net"
)

// InitTCP listen all tcp.bind and start accept connections.
func InitTCP() (err error) {
	var (
		listener *net.TCPListener
		addr     *net.TCPAddr
	)
	for _, bind := range Conf.TCPBind {
		if addr, err = net.ResolveTCPAddr("tcp4", bind); err != nil {
			log.Error("net.ResolveTCPAddr(\"tcp4\", \"%s\") error(%v)", bind, err)
			return
		}
		if listener, err = net.ListenTCP("tcp4", addr); err != nil {
			log.Error("net.ListenTCP(\"tcp4\", \"%s\") error(%v)", bind, err)
			return
		}
		// split N core accept
		for i := 0; i < Conf.MaxProc; i++ {
			log.Debug("start tcp accept[goroutine %d]: \"%s\"", i, bind)
			go DefaultServer.AcceptTCP(listener, i)
		}
	}
	return
}

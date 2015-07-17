package main

import (
	log "code.google.com/p/log4go"
	"net"
)

const (
	maxPackIntBuf = 4
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
		log.Debug("start tcp listen: \"%s\"", bind)
		// split N core accept
		for i := 0; i < Conf.MaxProc; i++ {
			go acceptTCP(DefaultServer, listener)
		}
	}
	return
}

// Accept accepts connections on the listener and serves requests
// for each incoming connection.  Accept blocks; the caller typically
// invokes it in a go statement.
func acceptTCP(server *Server, lis *net.TCPListener) {
	var (
		conn *net.TCPConn
		err  error
		r    int
	)
	for {
		if conn, err = lis.AcceptTCP(); err != nil {
			// if listener close then return
			log.Error("listener.Accept(\"%s\") error(%v)", lis.Addr().String(), err)
			return
		}
		if err = conn.SetKeepAlive(Conf.TCPKeepalive); err != nil {
			log.Error("conn.SetKeepAlive() error(%v)", err)
			return
		}
		if err = conn.SetReadBuffer(Conf.TCPSndbuf); err != nil {
			log.Error("conn.SetReadBuffer() error(%v)", err)
			return
		}
		if err = conn.SetWriteBuffer(Conf.TCPRcvbuf); err != nil {
			log.Error("conn.SetWriteBuffer() error(%v)", err)
			return
		}
		go serveTCP(server, conn, r)
		if r++; r == maxInt {
			r = 0
		}
	}
}

func serveTCP(server *Server, conn *net.TCPConn, r int) {
	var (
		// bufpool
		rrp = server.round.Reader(r) // reader
		wrp = server.round.Writer(r) // writer
		// timer
		tr = server.round.Timer(r)
		// buf
		rr = NewBufioReaderSize(rrp, conn, Conf.ReadBufSize)  // reader buf
		wr = NewBufioWriterSize(wrp, conn, Conf.WriteBufSize) // writer buf
		// ip addr
		lAddr = conn.LocalAddr().String()
		rAddr = conn.RemoteAddr().String()
	)
	log.Debug("start tcp serve \"%s\" with \"%s\"", lAddr, rAddr)
	server.serveTCP(conn, rr, wr, tr)
	PutBufioReader(rrp, rr)
	PutBufioWriter(wrp, wr)
}

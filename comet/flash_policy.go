package main

import (
	"goim/libs/bufio"
	"goim/libs/bytes"
	itime "goim/libs/time"
	"net"

	log "github.com/thinkboy/log4go"
)

const (
	FlashPolicyRequestLen = len("<policy-file-request/>")
)

var (
	FlashPolicyResponse []byte
)

// InitFlashPolicy listen all network interface and start accept connections.
func InitFlashPolicy() (err error) {
	var (
		listener *net.TCPListener
		addr     *net.TCPAddr
	)
	FlashPolicyResponse = []byte("<cross-domain-policy><allow-access-from domain=\"*\" to-ports=\"*\" /></cross-domain-policy>\n")
	FlashPolicyResponse = append(FlashPolicyResponse, 0)
	for _, bind := range Conf.FlashPolicyBind {
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
			go acceptFlashPolicy(DefaultServer, listener)
		}
	}
	return
}

// Accept accepts connections on the listener and serves requests
// for each incoming connection.  Accept blocks; the caller typically
// invokes it in a go statement.
func acceptFlashPolicy(server *Server, lis *net.TCPListener) {
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
		if err = conn.SetWriteBuffer(Conf.TCPRcvbuf); err != nil {
			log.Error("conn.SetWriteBuffer() error(%v)", err)
			return
		}
		go serveFlashPolicy(server, conn, r)
		if r++; r == maxInt {
			r = 0
		}
	}
}

func serveFlashPolicy(server *Server, conn *net.TCPConn, r int) {
	var (
		// timer
		tr = server.round.Timer(r)
		rp = server.round.Reader(r)
		wp = server.round.Writer(r)
		// ip addr
		lAddr = conn.LocalAddr().String()
		rAddr = conn.RemoteAddr().String()
	)
	if Debug {
		log.Debug("start tcp flash policy serve \"%s\" with \"%s\"", lAddr, rAddr)
	}

	server.serverFlashPolicy(conn, rp, wp, tr, rAddr)
}

func (server *Server) serverFlashPolicy(conn *net.TCPConn, rp, wp *bytes.Pool, tr *itime.Timer, rAddr string) {
	var (
		rr bufio.Reader
		wr bufio.Writer
		rb = rp.Get()
		wb = wp.Get()
	)
	rr.ResetBuffer(conn, rb.Bytes())
	wr.ResetBuffer(conn, wb.Bytes())

	trd := tr.Add(server.Options.HandshakeTimeout, func() { //安全协议设置超时时间
		conn.Close()
	})
	_, err := rr.Pop(FlashPolicyRequestLen)
	if err != nil {
		log.Error("rr.Pop() error(%v)", err)
		goto failed
	}
	_, err = wr.Write(FlashPolicyResponse)
	if err != nil {
		log.Error("wr.Write() error(%v)", err)
		goto failed
	}
	wr.Flush()
	if Conf.Debug {
		log.Debug("remote ip:%s write a flash safe proto succeed", rAddr)
	}

failed:
	tr.Del(trd)
	conn.Close()
	rp.Put(rb)
	wp.Put(wb)
}

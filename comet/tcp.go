package main

import (
	"goim/libs/bufio"
	"goim/libs/bytes"
	"goim/libs/define"
	"goim/libs/proto"
	itime "goim/libs/time"
	"io"
	"net"
	"time"

	log "github.com/thinkboy/log4go"
)

// InitTCP listen all tcp.bind and start accept connections.
func InitTCP(addrs []string, accept int) (err error) {
	var (
		bind     string
		listener *net.TCPListener
		addr     *net.TCPAddr
	)
	for _, bind = range addrs {
		if addr, err = net.ResolveTCPAddr("tcp4", bind); err != nil {
			log.Error("net.ResolveTCPAddr(\"tcp4\", \"%s\") error(%v)", bind, err)
			return
		}
		if listener, err = net.ListenTCP("tcp4", addr); err != nil {
			log.Error("net.ListenTCP(\"tcp4\", \"%s\") error(%v)", bind, err)
			return
		}
		if Debug {
			log.Debug("start tcp listen: \"%s\"", bind)
		}
		// split N core accept
		for i := 0; i < accept; i++ {
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
		if err = conn.SetKeepAlive(server.Options.TCPKeepalive); err != nil {
			log.Error("conn.SetKeepAlive() error(%v)", err)
			return
		}
		if err = conn.SetReadBuffer(server.Options.TCPRcvbuf); err != nil {
			log.Error("conn.SetReadBuffer() error(%v)", err)
			return
		}
		if err = conn.SetWriteBuffer(server.Options.TCPSndbuf); err != nil {
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
		// timer
		tr = server.round.Timer(r)
		rp = server.round.Reader(r)
		wp = server.round.Writer(r)
		// ip addr
		lAddr = conn.LocalAddr().String()
		rAddr = conn.RemoteAddr().String()
	)
	if Debug {
		log.Debug("start tcp serve \"%s\" with \"%s\"", lAddr, rAddr)
	}
	server.serveTCP(conn, rp, wp, tr)
}

// TODO linger close?
func (server *Server) serveTCP(conn *net.TCPConn, rp, wp *bytes.Pool, tr *itime.Timer) {
	var (
		err   error
		key   string
		white bool
		hb    time.Duration // heartbeat
		p     *proto.Proto
		b     *Bucket
		trd   *itime.TimerData
		rb    = rp.Get()
		wb    = wp.Get()
		ch    = NewChannel(server.Options.CliProto, server.Options.SvrProto, define.NoRoom)
		rr    = &ch.Reader
		wr    = &ch.Writer
	)
	ch.Reader.ResetBuffer(conn, rb.Bytes())
	ch.Writer.ResetBuffer(conn, wb.Bytes())
	// handshake
	trd = tr.Add(server.Options.HandshakeTimeout, func() {
		conn.Close()
	})
	// must not setadv, only used in auth
	if p, err = ch.CliProto.Set(); err == nil {
		if key, ch.RoomId, hb, err = server.authTCP(rr, wr, p); err == nil {
			b = server.Bucket(key)
			err = b.Put(key, ch)
		}
	}
	if err != nil {
		conn.Close()
		rp.Put(rb)
		wp.Put(wb)
		tr.Del(trd)
		log.Error("key: %s handshake failed error(%v)", key, err)
		return
	}
	trd.Key = key
	tr.Set(trd, hb)
	white = DefaultWhitelist.Contains(key)
	if white {
		DefaultWhitelist.Log.Printf("key: %s[%d] auth\n", key, ch.RoomId)
	}
	// hanshake ok start dispatch goroutine
	go server.dispatchTCP(key, conn, wr, wp, wb, ch)
	for {
		if p, err = ch.CliProto.Set(); err != nil {
			break
		}
		if white {
			DefaultWhitelist.Log.Printf("key: %s start read proto\n", key)
		}
		if err = p.ReadTCP(rr); err != nil {
			break
		}
		if white {
			DefaultWhitelist.Log.Printf("key: %s read proto:%v\n", key, p)
		}
		if p.Operation == define.OP_HEARTBEAT {
			tr.Set(trd, hb)
			p.Body = nil
			p.Operation = define.OP_HEARTBEAT_REPLY
			if Debug {
				log.Debug("key: %s receive heartbeat", key)
			}
		} else {
			if err = server.operator.Operate(p); err != nil {
				break
			}
		}
		if white {
			DefaultWhitelist.Log.Printf("key: %s process proto:%v\n", key, p)
		}
		ch.CliProto.SetAdv()
		ch.Signal()
		if white {
			DefaultWhitelist.Log.Printf("key: %s signal\n", key)
		}
	}
	if white {
		DefaultWhitelist.Log.Printf("key: %s server tcp error(%v)\n", key, err)
	}
	if err != nil && err != io.EOF {
		log.Error("key: %s server tcp failed error(%v)", key, err)
	}
	b.Del(key)
	tr.Del(trd)
	rp.Put(rb)
	conn.Close()
	ch.Close()
	if err = server.operator.Disconnect(key, ch.RoomId); err != nil {
		log.Error("key: %s operator do disconnect error(%v)", key, err)
	}
	if white {
		DefaultWhitelist.Log.Printf("key: %s disconnect error(%v)\n", key, err)
	}
	if Debug {
		log.Debug("key: %s server tcp goroutine exit", key)
	}
	return
}

// dispatch accepts connections on the listener and serves requests
// for each incoming connection.  dispatch blocks; the caller typically
// invokes it in a go statement.
func (server *Server) dispatchTCP(key string, conn *net.TCPConn, wr *bufio.Writer, wp *bytes.Pool, wb *bytes.Buffer, ch *Channel) {
	var (
		err    error
		finish bool
		white  = DefaultWhitelist.Contains(key)
	)
	if Debug {
		log.Debug("key: %s start dispatch tcp goroutine", key)
	}
	for {
		if white {
			DefaultWhitelist.Log.Printf("key: %s wait proto ready\n", key)
		}
		var p = ch.Ready()
		if white {
			DefaultWhitelist.Log.Printf("key: %s proto ready\n", key)
		}
		if Debug {
			log.Debug("key:%s dispatch msg:%v", key, *p)
		}
		switch p {
		case proto.ProtoFinish:
			if white {
				DefaultWhitelist.Log.Printf("key: %s receive proto finish\n", key)
			}
			if Debug {
				log.Debug("key: %s wakeup exit dispatch goroutine", key)
			}
			finish = true
			goto failed
		case proto.ProtoReady:
			// fetch message from svrbox(client send)
			for {
				if p, err = ch.CliProto.Get(); err != nil {
					err = nil // must be empty error
					break
				}
				if white {
					DefaultWhitelist.Log.Printf("key: %s start write client proto%v\n", key, p)
				}
				if err = p.WriteTCP(wr); err != nil {
					goto failed
				}
				if white {
					DefaultWhitelist.Log.Printf("key: %s write client proto%v\n", key, p)
				}
				p.Body = nil // avoid memory leak
				ch.CliProto.GetAdv()
			}
		default:
			if white {
				DefaultWhitelist.Log.Printf("key: %s start write server proto%v\n", key, p)
			}
			// server send
			if err = p.WriteTCP(wr); err != nil {
				goto failed
			}
			if white {
				DefaultWhitelist.Log.Printf("key: %s write server proto%v\n", key, p)
			}
		}
		if white {
			DefaultWhitelist.Log.Printf("key: %s start flush \n", key)
		}
		// only hungry flush response
		if err = wr.Flush(); err != nil {
			break
		}
		if white {
			DefaultWhitelist.Log.Printf("key: %s flush\n", key)
		}
	}
failed:
	if white {
		DefaultWhitelist.Log.Printf("key: dispatch tcp error(%v)\n", key, err)
	}
	if err != nil {
		log.Error("key: %s dispatch tcp error(%v)", key, err)
	}
	conn.Close()
	wp.Put(wb)
	// must ensure all channel message discard, for reader won't blocking Signal
	for !finish {
		finish = (ch.Ready() == proto.ProtoFinish)
	}
	if Debug {
		log.Debug("key: %s dispatch goroutine exit", key)
	}
	return
}

// auth for goim handshake with client, use rsa & aes.
func (server *Server) authTCP(rr *bufio.Reader, wr *bufio.Writer, p *proto.Proto) (key string, rid int32, heartbeat time.Duration, err error) {
	if err = p.ReadTCP(rr); err != nil {
		return
	}
	if p.Operation != define.OP_AUTH {
		log.Warn("auth operation not valid: %d", p.Operation)
		err = ErrOperation
		return
	}
	if key, rid, heartbeat, err = server.operator.Connect(p); err != nil {
		return
	}
	p.Body = nil
	p.Operation = define.OP_AUTH_REPLY
	if err = p.WriteTCP(wr); err != nil {
		return
	}
	err = wr.Flush()
	return
}

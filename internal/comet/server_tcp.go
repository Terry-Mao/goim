package comet

import (
	"context"
	"io"
	"net"
	"strings"
	"time"

	"github.com/Terry-Mao/goim/api/comet/grpc"
	"github.com/Terry-Mao/goim/internal/comet/conf"
	"github.com/Terry-Mao/goim/pkg/bufio"
	"github.com/Terry-Mao/goim/pkg/bytes"
	xtime "github.com/Terry-Mao/goim/pkg/time"
	log "github.com/golang/glog"
)

const (
	maxInt = 1<<31 - 1
)

// InitTCP listen all tcp.bind and start accept connections.
func InitTCP(server *Server, addrs []string, accept int) (err error) {
	var (
		bind     string
		listener *net.TCPListener
		addr     *net.TCPAddr
	)
	// 監聽多個Tcp Port
	for _, bind = range addrs {
		if addr, err = net.ResolveTCPAddr("tcp", bind); err != nil {
			log.Errorf("net.ResolveTCPAddr(tcp, %s) error(%v)", bind, err)
			return
		}
		if listener, err = net.ListenTCP("tcp", addr); err != nil {
			log.Errorf("net.ListenTCP(tcp, %s) error(%v)", bind, err)
			return
		}
		log.Infof("start tcp listen: %s", bind)

		// 一個Tcp Port根據CPU核心數開goroutine監聽Tcp
		for i := 0; i < accept; i++ {
			go acceptTCP(server, listener)
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

		// 取Pool的rand seed
		r int
	)
	for {
		// tcp監聽並連線
		if conn, err = lis.AcceptTCP(); err != nil {
			log.Errorf("listener.Accept(\"%s\") error(%v)", lis.Addr().String(), err)
			return
		}
		// tcp 開啟KeepAlive
		if err = conn.SetKeepAlive(server.c.TCP.KeepAlive); err != nil {
			log.Errorf("conn.SetKeepAlive() error(%v)", err)
			return
		}
		// tcp讀取資料的緩衝區大小，該緩衝區為0時會阻塞，此值通常設定完後，系統會自行在多一倍，設定1024會變2304
		if err = conn.SetReadBuffer(server.c.TCP.Rcvbuf); err != nil {
			log.Errorf("conn.SetReadBuffer() error(%v)", err)
			return
		}
		// tcp寫資料的緩衝區大小，該緩衝區滿到無法發送時會阻塞，此值通常設定完後系統會自行在多一倍，設定1024會變2304
		if err = conn.SetWriteBuffer(server.c.TCP.Sndbuf); err != nil {
			log.Errorf("conn.SetWriteBuffer() error(%v)", err)
			return
		}

		go serveTCP(server, conn, r)
		if r++; r == maxInt {
			r = 0
		}
	}
}

// tcp連線後的邏輯處理
func serveTCP(s *Server, conn *net.TCPConn, r int) {
	var (
		// 任務倒數計時器
		tr = s.round.Timer(r)

		// Reader Buffer
		rp = s.round.Reader(r)

		// Writer Buffer
		wp = s.round.Writer(r)

		// 本地ip:port
		lAddr = conn.LocalAddr().String()

		// tcp來源端ip:port
		rAddr = conn.RemoteAddr().String()
	)
	if conf.Conf.Debug {
		log.Infof("start tcp serve \"%s\" with \"%s\"", lAddr, rAddr)
	}
	s.ServeTCP(conn, rp, wp, tr)
}

// ServeTCP serve a tcp connection.
func (s *Server) ServeTCP(conn *net.TCPConn, rp, wp *bytes.Pool, tr *xtime.Timer) {
	var (
		err error

		// 房間id
		rid string

		// tcp 連線的tag，可以用於設置訊息推送條件
		accepts []int32

		// 心跳時間週期
		hb time.Duration

		//
		white bool

		// grpc 自訂Protocol
		p *grpc.Proto

		// 管理Channel與Room
		b *Bucket

		// 時間倒數任務
		trd *xtime.TimerData

		// 現在時間
		lastHb = time.Now()

		// 用於讀的Buffer
		rb = rp.Get()

		// 用於寫的Buffer
		wb = wp.Get()

		// 此tcp連線的Channel
		ch = NewChannel(s.c.Protocol.CliProto, s.c.Protocol.SvrProto)

		// Reader byte
		rr = &ch.Reader

		// Writer byte
		wr = &ch.Writer
	)

	// Channel設置的讀寫Buffer(由Pool取得之後會還給Pool做復用)
	ch.Reader.ResetBuffer(conn, rb.Bytes())
	ch.Writer.ResetBuffer(conn, wb.Bytes())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	step := 0

	// 心跳超時後的邏輯
	trd = tr.Add(time.Duration(s.c.Protocol.HandshakeTimeout), func() {
		conn.Close()
		log.Errorf("key: %s remoteIP: %s step: %d tcp handshake timeout", ch.Key, conn.RemoteAddr().String(), step)
	})

	ch.IP, _, _ = net.SplitHostPort(conn.RemoteAddr().String())
	step = 1

	// tcp連線做auth
	if p, err = ch.CliProto.Set(); err == nil {
		if ch.Mid, ch.Key, rid, accepts, hb, err = s.authTCP(ctx, rr, wr, p); err == nil {
			// 將user Channel放到某一個Bucket內做保存
			ch.Watch(accepts...)
			b = s.Bucket(ch.Key)
			err = b.Put(rid, ch)
			if conf.Conf.Debug {
				log.Infof("tcp connnected key:%s mid:%d proto:%+v", ch.Key, ch.Mid, p)
			}
		}
	}

	step = 2

	// 如果操作有異常則
	// 1. 回收讀寫Buffer
	// 2. 解除心跳任務(close 連線)
	// 3. 關閉連線
	if err != nil {
		conn.Close()
		rp.Put(rb)
		wp.Put(wb)
		tr.Del(trd)
		log.Errorf("key: %s handshake failed error(%v)", ch.Key, err)
		return
	}

	// 進入某房間成功後重置心跳任務
	trd.Key = ch.Key
	tr.Set(trd, hb)

	white = whitelist.Contains(ch.Mid)
	if white {
		whitelist.Printf("key: %s[%s] auth\n", ch.Key, rid)
	}

	step = 3

	// 處理訊息推送
	go s.dispatchTCP(conn, wr, wp, wb, ch)

	serverHeartbeat := s.RandServerHearbeat()

	// 處理tcp送過來的資料邏輯
	for {
		if p, err = ch.CliProto.Set(); err != nil {
			break
		}
		if white {
			whitelist.Printf("key: %s start read proto\n", ch.Key)
		}
		if err = p.ReadTCP(rr); err != nil {
			break
		}
		if white {
			whitelist.Printf("key: %s read proto:%v\n", ch.Key, p)
		}

		// client回應心跳，server要回覆心跳結果
		if p.Op == grpc.OpHeartbeat {
			tr.Set(trd, hb)
			p.Op = grpc.OpHeartbeatReply
			p.Body = nil
			// NOTE: send server heartbeat for a long time
			if now := time.Now(); now.Sub(lastHb) > serverHeartbeat {
				if err1 := s.Heartbeat(ctx, ch.Mid, ch.Key); err1 == nil {
					lastHb = now
				}
			}
			if conf.Conf.Debug {
				log.Infof("tcp heartbeat receive key:%s, mid:%d", ch.Key, ch.Mid)
			}
			step++
		} else {
			// 非心跳動作
			if err = s.Operate(ctx, p, ch, b); err != nil {
				break
			}
		}
		if white {
			whitelist.Printf("key: %s process proto:%v\n", ch.Key, p)
		}

		// 寫的游標要++讓Get可以取得已寫入的Proto
		ch.CliProto.SetAdv()

		// 通知負責訊息推播goroutine處理本次接收到的資料
		ch.Signal()

		if white {
			whitelist.Printf("key: %s signal\n", ch.Key)
		}
	}

	// 如果某人連線有異常或是server要踢人則
	// 1. 從Bucket移除user Channel，這樣對Bucket內的Channel才都是活人
	// 2. 解除心跳任務(close 連線)
	// 3. 回收讀Buffer，不回收寫的Buffer是因為Channel close後dispatchTCP會被通知到並回收寫的Buffer
	// 4. 關閉連線
	// 5. 通知logic某人下線了
	if white {
		whitelist.Printf("key: %s server tcp error(%v)\n", ch.Key, err)
	}
	if err != nil && err != io.EOF && !strings.Contains(err.Error(), "closed") {
		log.Errorf("key: %s server tcp failed error(%v)", ch.Key, err)
	}
	b.Del(ch)
	tr.Del(trd)
	rp.Put(rb)
	conn.Close()
	ch.Close()
	if err = s.Disconnect(ctx, ch.Mid, ch.Key); err != nil {
		log.Errorf("key: %s mid: %d operator do disconnect error(%v)", ch.Key, ch.Mid, err)
	}
	if white {
		whitelist.Printf("key: %s mid: %d disconnect error(%v)\n", ch.Key, ch.Mid, err)
	}
	if conf.Conf.Debug {
		log.Infof("tcp disconnected key: %s mid: %d", ch.Key, ch.Mid)
	}
}

// dispatch accepts connections on the listener and serves requests
// for each incoming connection.  dispatch blocks; the caller typically
// invokes it in a go statement.
func (s *Server) dispatchTCP(conn *net.TCPConn, wr *bufio.Writer, wp *bytes.Pool, wb *bytes.Buffer, ch *Channel) {
	var (
		err    error
		finish bool
		online int32
		white  = whitelist.Contains(ch.Mid)
	)
	if conf.Conf.Debug {
		log.Infof("key: %s start dispatch tcp goroutine", ch.Key)
	}
	for {
		if white {
			whitelist.Printf("key: %s wait proto ready\n", ch.Key)
		}

		// 接收到別人通知説有資料要推送，沒資料時就阻塞
		var p = ch.Ready()
		if white {
			whitelist.Printf("key: %s proto ready\n", ch.Key)
		}
		if conf.Conf.Debug {
			log.Infof("key:%s dispatch msg:%v", ch.Key, *p)
		}
		switch p {

		// tcp連線要關閉
		case grpc.ProtoFinish:
			if white {
				whitelist.Printf("key: %s receive proto finish\n", ch.Key)
			}
			if conf.Conf.Debug {
				log.Infof("key: %s wakeup exit dispatch goroutine", ch.Key)
			}
			finish = true
			goto failed

			// 有資料需要推送
		case grpc.ProtoReady:
			for {
				// 取得上次透過Set()寫入資料的Proto
				if p, err = ch.CliProto.Get(); err != nil {
					break
				}
				if white {
					whitelist.Printf("key: %s start write client proto%v\n", ch.Key, p)
				}
				if p.Op == grpc.OpHeartbeatReply {
					if ch.Room != nil {
						online = ch.Room.OnlineNum()
					}
					if err = p.WriteTCPHeart(wr, online); err != nil {
						goto failed
					}
				} else {
					if err = p.WriteTCP(wr); err != nil {
						goto failed
					}
				}
				if white {
					whitelist.Printf("key: %s write client proto%v\n", ch.Key, p)
				}
				p.Body = nil

				// 讀的游標++
				ch.CliProto.GetAdv()
			}
		default:
			if white {
				whitelist.Printf("key: %s start write server proto%v\n", ch.Key, p)
			}
			// server send
			if err = p.WriteTCP(wr); err != nil {
				goto failed
			}
			if white {
				whitelist.Printf("key: %s write server proto%v\n", ch.Key, p)
			}
			if conf.Conf.Debug {
				log.Infof("tcp sent a message key:%s mid:%d proto:%+v", ch.Key, ch.Mid, p)
			}
		}
		if white {
			whitelist.Printf("key: %s start flush \n", ch.Key)
		}
		// 送出資料給client
		if err = wr.Flush(); err != nil {
			break
		}
		if white {
			whitelist.Printf("key: %s flush\n", ch.Key)
		}
	}
	// 連線有異常或是server要踢人
	// 1. 連線close
	// 2. 回收寫的Buffter
failed:
	if white {
		whitelist.Printf("key: %s dispatch tcp error(%v)\n", ch.Key, err)
	}
	if err != nil {
		log.Errorf("key: %s dispatch tcp error(%v)", ch.Key, err)
	}
	conn.Close()
	wp.Put(wb)
	// must ensure all channel message discard, for reader won't blocking Signal
	for !finish {
		finish = (ch.Ready() == grpc.ProtoFinish)
	}
	if conf.Conf.Debug {
		log.Infof("key: %s dispatch goroutine exit", ch.Key)
	}
}

// auth for goim handshake with client, use rsa & aes.
func (s *Server) authTCP(ctx context.Context, rr *bufio.Reader, wr *bufio.Writer, p *grpc.Proto) (mid int64, key, rid string, accepts []int32, hb time.Duration, err error) {
	for {
		if err = p.ReadTCP(rr); err != nil {
			return
		}
		// 如果第一次連線送的資料不是請求連接到某房間則會一直等待
		if p.Op == grpc.OpAuth {
			break
		} else {
			log.Errorf("tcp request operation(%d) not auth", p.Op)
		}
	}
	if mid, key, rid, accepts, hb, err = s.Connect(ctx, p, ""); err != nil {
		log.Errorf("authTCP.Connect(key:%v).err(%v)", key, err)
		return
	}

	// 回覆連線至某房間結果
	p.Op = grpc.OpAuthReply
	p.Body = nil
	if err = p.WriteTCP(wr); err != nil {
		log.Errorf("authTCP.WriteTCP(key:%v).err(%v)", key, err)
		return
	}
	err = wr.Flush()
	return
}

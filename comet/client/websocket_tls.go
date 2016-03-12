package main

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"time"

	log "github.com/thinkboy/log4go"
	"golang.org/x/net/websocket"
)

func initWebsocketTLS() {
	origin := "https://" + Conf.WebsocketAddr + "/sub"
	url := "wss://" + Conf.WebsocketAddr + "/sub"
	conf, err := websocket.NewConfig(url, origin)
	if err != nil {
		log.Error("websocket.NewConfig(\"%s\") error(%v)", Conf.WebsocketAddr, err)
		return
	}
	roots := x509.NewCertPool()
	certPem, err := ioutil.ReadFile(Conf.CertFile)
	if err != nil {
		panic(err)
	}
	ok := roots.AppendCertsFromPEM(certPem)
	if !ok {
		panic("failed to parse root certificate")
	}

	tlsConf := &tls.Config{
		//InsecureSkipVerify: true,
		RootCAs:    roots,
		ServerName: "bili.com",
	}
	conf.TlsConfig = tlsConf

	conn, err := websocket.DialConfig(conf)
	if err != nil {
		log.Error("websocket.Dial(\"%s\") error(%v)", Conf.WebsocketAddr, err)
		return
	}

	proto := new(Proto)
	proto.Ver = 1
	// auth
	// test handshake timeout
	// time.Sleep(time.Second * 31)
	proto.Operation = OP_AUTH
	seqId := int32(0)
	proto.SeqId = seqId
	proto.Body = []byte("{\"test\":1}")
	if err = websocketWriteProto(conn, proto); err != nil {
		log.Error("websocketWriteProto() error(%v)", err)
		return
	}
	if err = websocketReadProto(conn, proto); err != nil {
		log.Error("websocketReadProto() error(%v)", err)
		return
	}
	log.Debug("auth ok, proto: %v", proto)
	seqId++
	// writer
	go func() {
		proto1 := new(Proto)
		for {
			// heartbeat
			proto1.Operation = OP_HEARTBEAT
			proto1.SeqId = seqId
			proto1.Body = nil
			if err = websocketWriteProto(conn, proto1); err != nil {
				log.Error("tcpWriteProto() error(%v)", err)
				return
			}
			// test heartbeat
			//time.Sleep(time.Second * 31)
			seqId++
			// op_test
			proto1.Operation = OP_TEST
			proto1.SeqId = seqId
			if err = websocketWriteProto(conn, proto1); err != nil {
				log.Error("tcpWriteProto() error(%v)", err)
				return
			}
			seqId++
			time.Sleep(10000 * time.Millisecond)
		}
	}()
	// reader
	for {
		if err = websocketReadProto(conn, proto); err != nil {
			log.Error("tcpReadProto() error(%v)", err)
			return
		}
		if proto.Operation == OP_HEARTBEAT_REPLY {
			log.Debug("receive heartbeat")
			if err = conn.SetReadDeadline(time.Now().Add(25 * time.Second)); err != nil {
				log.Error("conn.SetReadDeadline() error(%v)", err)
				return
			}
		} else if proto.Operation == OP_TEST_REPLY {
			log.Debug("body: %s", string(proto.Body))
		} else if proto.Operation == OP_SEND_SMS_REPLY {
			log.Debug("body: %s", string(proto.Body))
		}
	}
}

package main

import (
	"encoding/json"
	"time"

	log "github.com/thinkboy/log4go"
	"golang.org/x/net/websocket"
)

func initWebsocket() {
	origin := "http://" + Conf.WebsocketAddr + "/sub"
	url := "ws://" + Conf.WebsocketAddr + "/sub"
	conn, err := websocket.Dial(url, "", origin)
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

func websocketReadProto(conn *websocket.Conn, p *Proto) error {
	msg, _ := json.Marshal(p)
	log.Debug("%s", string(msg))
	return websocket.JSON.Receive(conn, p)
}

func websocketWriteProto(conn *websocket.Conn, p *Proto) error {
	if p.Body == nil {
		p.Body = []byte("{}")
	}
	return websocket.JSON.Send(conn, p)
}

package websocket

import (
	"net"
	"reflect"
	"testing"
	"time"

	"golang.org/x/net/websocket"

	"github.com/Terry-Mao/goim/pkg/bufio"
)

func TestServer(t *testing.T) {
	var (
		data = []byte{0, 1, 2}
	)
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		t.FailNow()
	}
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			t.Error(err)
		}
		rd := bufio.NewReader(conn)
		wr := bufio.NewWriter(conn)
		req, err := ReadRequest(rd)
		if err != nil {
			t.Error(err)
		}
		if req.RequestURI != "/sub" {
			t.Error(err)
		}
		ws, err := Upgrade(conn, rd, wr, req)
		if err != nil {
			t.Error(err)
		}
		if err = ws.WriteMessage(BinaryMessage, data); err != nil {
			t.Error(err)
		}
		if err = ws.Flush(); err != nil {
			t.Error(err)
		}
		op, b, err := ws.ReadMessage()
		if err != nil || op != BinaryMessage || !reflect.DeepEqual(b, data) {
			t.Error(err)
		}
	}()
	time.Sleep(time.Millisecond * 100)
	// ws client
	ws, err := websocket.Dial("ws://127.0.0.1:8080/sub", "", "*")
	if err != nil {
		t.FailNow()
	}
	// receive binary frame
	var b []byte
	if err = websocket.Message.Receive(ws, &b); err != nil {
		t.FailNow()
	}
	if !reflect.DeepEqual(b, data) {
		t.FailNow()
	}
	// send binary frame
	if err = websocket.Message.Send(ws, data); err != nil {
		t.FailNow()
	}
}

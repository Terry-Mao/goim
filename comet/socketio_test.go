package main

import (
	"net/http"
	"testing"
	log "github.com/thinkboy/log4go"
	"github.com/nkovacs/go-socket.io"
)

func TestRouterGet(t *testing.T) {

	server, err := socketio.NewServer(nil)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}


	server.On("connection", func(so socketio.Socket) {
		log.Info("on connection")
		so.Join("chat")
		so.On("chat message", func(msg string) {
			log.Info("emit:", so.Emit("chat message", msg))
			so.BroadcastTo("chat", "chat message", msg)
		})
		so.On("disconnection", func() {
			log.Info("on disconnect")
		})
	})
	server.On("error", func(so socketio.Socket, err error) {
		log.Info("error:", err)
	})

	http.Handle("/socket.io/", server)
	http.Handle("/", http.FileServer(http.Dir("./public")))
	log.Info("Serving at localhost:12345...")
	err = http.ListenAndServe(":12345", nil)

	if err != nil {
		t.Error(err)
		t.FailNow()
	}

}

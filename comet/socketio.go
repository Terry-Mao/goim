package main

import (
	"goim/libs/proto"
	"time"
	"net/http"

	log "github.com/thinkboy/log4go"
	"github.com/nkovacs/go-socket.io"

)

// Initsocket.io listen all tcp.bind and start accept connections.
func InitSocketIO(addrs []string, transport []string, accept string) (err error) {
	var bind string
	for _, bind = range addrs {
		server,err := socketio.NewServer(transport)
	    if(err != nil){
		    log.Warn("socketio init err")
	    }
		http.Handle("/socket.io/", server)
	    log.Info("socketio Serving at ",bind)
	    go http.ListenAndServe(bind, nil)
	    key := accept
	    go dispatchSocketIOEvent(server,key)
		}
	return err
}

// dispatch accepts connections on the listener and serves requests
// for each incoming connection.  dispatch blocks; the caller typically
// invokes it in a go statement.
func dispatchSocketIOEvent(server *socketio.Server,key string){
	if Debug {
		log.Debug("key: %s start dispatch tcp goroutine", key)
	}
	server.On("connection", func(so socketio.Socket) {
		log.Info("on connection")
		so.Join(key)
		so.On("chat message", func(msg string) {
			log.Info("emit:", so.Emit("chat message", msg))
			so.BroadcastTo(key, "chat message", msg)
		})
		so.On("disconnection", func() {
			log.Info("on disconnect")
		})
	})
	server.On("error", func(so socketio.Socket, err error) {
		log.Info("error:", err)
	})
	if Debug {
		log.Debug("key: %s dispatch goroutine exit", key)
	}
	return
}

// auth for goim handshake with client, use rsa & aes.
func (server *Server) authSocketio(ws *socketio.Socket, p *proto.Proto) (key string, rid int32, heartbeat time.Duration, err error) {
	if err = p.ReadWebsocket(ws); err != nil {
		return
	}
	return
}

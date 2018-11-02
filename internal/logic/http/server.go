package http

import (
	"net/http"
	"time"

	"github.com/Terry-Mao/goim/internal/logic"
	"github.com/Terry-Mao/goim/internal/logic/conf"
)

// Server is http server.
type Server struct {
	conf  *conf.HTTPServer
	logic *logic.Logic
}

// New new a http server.
func New(c *conf.HTTPServer, l *logic.Logic) *http.Server {
	s := &Server{
		conf:  c,
		logic: l,
	}
	srv := &http.Server{
		Addr:           c.Addr,
		Handler:        s.newHTTPServeMux(),
		ReadTimeout:    time.Duration(c.ReadTimeout),
		WriteTimeout:   time.Duration(c.WriteTimeout),
		MaxHeaderBytes: 1 << 20,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()
	return srv
}

func (s *Server) newHTTPServeMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/goim/push/keys", s.pushKeys)
	mux.HandleFunc("/goim/push/mids", s.pushMids)
	mux.HandleFunc("/goim/push/room", s.pushRoom)
	mux.HandleFunc("/goim/push/all", s.pushAll)
	mux.HandleFunc("/goim/online/top", s.onlineTop)
	mux.HandleFunc("/goim/online/room", s.onlineRoom)
	mux.HandleFunc("/goim/online/total", s.onlineTotal)
	mux.HandleFunc("/goim/nodes/weighted", s.nodesWeighted)
	mux.HandleFunc("/goim/nodes/debug", s.nodesDebug)
	mux.HandleFunc("/goim/nodes/infos", s.nodesInfos)
	return mux
}

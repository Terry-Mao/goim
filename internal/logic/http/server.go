package http

import (
	"github.com/Terry-Mao/goim/internal/logic"
	"github.com/Terry-Mao/goim/internal/logic/conf"

	"github.com/gin-gonic/gin"
)

// Server is http server.
type Server struct {
	engine *gin.Engine
	logic  *logic.Logic
}

// New new a http server.
func New(c *conf.HTTPServer, l *logic.Logic) *Server {
	engine := gin.New()
	engine.Use(loggerHandler, recoverHandler)
	go func() {
		if err := engine.Run(c.Addr); err != nil {
			panic(err)
		}
	}()
	s := &Server{
		engine: engine,
		logic:  l,
	}
	s.initRouter()
	return s
}

func (s *Server) initRouter() {
	group := s.engine.Group("/goim")
	group.POST("/goim/push/keys", s.pushKeys)
	group.POST("/goim/push/mids", s.pushMids)
	group.POST("/goim/push/room", s.pushRoom)
	group.POST("/goim/push/all", s.pushAll)
	group.GET("/goim/online/top", s.onlineTop)
	group.GET("/goim/online/room", s.onlineRoom)
	group.GET("/goim/online/total", s.onlineTotal)
	group.GET("/goim/nodes/weighted", s.nodesWeighted)
	group.GET("/goim/nodes/debug", s.nodesDebug)
	group.GET("/goim/nodes/infos", s.nodesInfos)
	return
}

// Close close the server.
func (s *Server) Close() {

}

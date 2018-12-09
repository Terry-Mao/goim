package http

import (
	"context"
	"net/http"
)

func (s *Server) nodesWeighted(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	platStr := query.Get("platform")
	res := s.logic.NodesWeighted(context.TODO(), platStr, r.RemoteAddr)
	writeJSON(w, OK, res)
}

func (s *Server) nodesDebug(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	ipStr := query.Get("ip")
	nodes, region, province, err := s.logic.NodesDebug(context.TODO(), ipStr)
	if err != nil {
		writeJSON(w, ServerErr, nil)
		return
	}
	res := make(map[string]interface{})
	res["nodes"] = nodes
	res["region"] = region
	res["province"] = province
	writeJSON(w, OK, res)
}

func (s *Server) nodesInfos(w http.ResponseWriter, r *http.Request) {
	res, err := s.logic.NodesInfos(context.TODO())
	if err != nil {
		writeJSON(w, ServerErr, nil)
		return
	}
	writeJSON(w, OK, res)
}

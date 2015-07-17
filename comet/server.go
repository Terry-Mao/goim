package main

import (
	log "code.google.com/p/log4go"
	"github.com/Terry-Mao/goim/libs/hash/cityhash"
)

var (
	maxInt        = 1<<31 - 1
	emptyJSONBody = []byte("{}")
)

type Server struct {
	buckets   []*Bucket // subkey bucket
	bucketIdx uint32
	round     *Round // accept round store
	operator  Operator
}

// NewServer returns a new Server.
func NewServer(b []*Bucket, r *Round, o Operator) *Server {
	s := new(Server)
	s.buckets = b
	s.bucketIdx = uint32(len(b))
	s.round = r
	s.operator = o
	return s
}

func (server *Server) Bucket(subKey string) *Bucket {
	idx := cityhash.CityHash32([]byte(subKey), uint32(len(subKey))) % server.bucketIdx
	log.Debug("\"%s\" hit channel bucket index: %d use cityhash", subKey, idx)
	return server.buckets[idx]
}

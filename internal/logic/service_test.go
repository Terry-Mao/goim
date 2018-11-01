package service

import (
	"flag"
	"testing"

	"github.com/Terry-Mao/goim/internal/logic/conf"
)

var (
	s *Service
)

func TestMain(m *testing.M) {
	flag.Set("conf", "../../cmd/logic/logic-example.toml")
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	s = New(conf.Conf)
	m.Run()
}

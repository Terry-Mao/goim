package logic

import (
	"flag"
	"testing"

	"github.com/Terry-Mao/goim/internal/logic/conf"
)

var (
	l *Logic
)

func TestMain(m *testing.M) {
	flag.Set("conf", "../../cmd/logic/logic-example.toml")
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	l = New(conf.Conf)
	m.Run()
}

package logic

import (
	"flag"
	"os"
	"testing"

	"github.com/Terry-Mao/goim/internal/logic/conf"
)

var (
	l *Logic
)

func TestMain(m *testing.M) {
	if err := flag.Set("conf", "../../cmd/logic/logic.toml"); err != nil {
		panic(err)
	}
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	l = New(conf.Conf)
	os.Exit(m.Run())
}

package dao

import (
	"flag"
	"os"
	"testing"

	"github.com/Terry-Mao/goim/internal/logic/conf"
)

var (
	d *Dao
)

func TestMain(m *testing.M) {
	if err := flag.Set("conf", "../../../cmd/logic/logic.toml"); err != nil {
		panic(err)
	}
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	d = New(conf.Conf)
	os.Exit(m.Run())
}

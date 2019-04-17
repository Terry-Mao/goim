package nats

import (
	"context"
	"os"
	"strconv"
	"testing"

	"github.com/Terry-Mao/goim/internal/logic/conf"
)

var (
	d *Dao
)

func TestMain(m *testing.M) {
	c := conf.Default()

	d = New(c)
	os.Exit(m.Run())
}

func TestPushMsg(t *testing.T) {
	_ = d.PushMsg(context.TODO(), 122, "room111", []string{"test", "tttt"}, []byte("test"))
}

func BenchmarkDao_PushMsg(b *testing.B) {
	// b.StopTimer()
	//
	// b.StartTimer()
	for n := 0; n < b.N; n++ {
		_ = d.PushMsg(context.TODO(), 122, "room111", []string{strconv.Itoa(n), "tttt"}, []byte("test"))
	}
}

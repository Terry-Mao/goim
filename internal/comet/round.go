package comet

import (
	"github.com/Terry-Mao/goim/internal/comet/conf"
	"github.com/Terry-Mao/goim/pkg/bytes"
	"github.com/Terry-Mao/goim/pkg/time"
)

// RoundOptions round options.
type RoundOptions struct {
	// 每次要分配多少個用於time.Timer的Pool
	Timer int

	// 每個time.Timer一開始能接收的TimerData數量
	TimerSize int

	// 每次要分配多少個用於Reader bytes的Pool
	Reader int

	// 每個Reader bytes Pool有多少個Buffer
	ReadBuf int

	// 每個Reader bytes Pool的Buffer能有多大的空間
	ReadBufSize int

	// 每次要分配多少個用於Writer bytes的Pool
	Writer int

	// 每個Writer bytes Pool有多少個Buffer
	WriteBuf int

	// 每個Writer bytes Pool的Buffer能有多大的空間
	WriteBufSize int
}

// 管理(Reader and Writer bytes) and Timer Pool
type Round struct {
	// 管理Reader bytes Pool
	readers []bytes.Pool

	// 管理Writer bytes Pool
	writers []bytes.Pool

	// 管理Timer Pool
	timers []time.Timer

	// Pool相關config
	options RoundOptions
}

// 1. 為了優化內存所以自行設計一個bytes pool來管理
// 2. 當一個client連線成功後需要有一個心跳機制去維護連線可不可用
//    所以每一個client需要搭配一個倒數計時器，如果有100w+連線就要100w+ time.NewTicker
//    為了優化這一塊所以自行實現一個time
func NewRound(c *conf.Config) (r *Round) {
	var i int
	r = &Round{
		options: RoundOptions{
			Reader:       c.TCP.Reader,
			ReadBuf:      c.TCP.ReadBuf,
			ReadBufSize:  c.TCP.ReadBufSize,
			Writer:       c.TCP.Writer,
			WriteBuf:     c.TCP.WriteBuf,
			WriteBufSize: c.TCP.WriteBufSize,
			Timer:        c.Protocol.Timer,
			TimerSize:    c.Protocol.TimerSize,
		}}

	// 依照config的設定初始化
	// 1.決定一開始有多少個Reader bytes.Pool，每個Pool有多少個Buffer
	//   每個Buffer管理[]byte容量是多少
	// 2.決定一開始有多少個Writer bytes.Pool，每個Pool有多少個Buffer
	//   每個Buffer管理[]byte容量是多少
	// 3. 決定一開始有多少個time.Timer，每個Timer能容納多少個TimerData
	r.readers = make([]bytes.Pool, r.options.Reader)
	for i = 0; i < r.options.Reader; i++ {
		r.readers[i].Init(r.options.ReadBuf, r.options.ReadBufSize)
	}

	r.writers = make([]bytes.Pool, r.options.Writer)
	for i = 0; i < r.options.Writer; i++ {
		r.writers[i].Init(r.options.WriteBuf, r.options.WriteBufSize)
	}

	r.timers = make([]time.Timer, r.options.Timer)
	for i = 0; i < r.options.Timer; i++ {
		r.timers[i].Init(r.options.TimerSize)
	}
	return
}

// 取Timer Pool，給定某數字以取餘數方式給其中一個，用於分散鎖競爭增加併發量
func (r *Round) Timer(rn int) *time.Timer {
	return &(r.timers[rn%r.options.Timer])
}

// 取Reader Pool，給定某數字以取餘數方式給其中一個，用於分散鎖競爭增加併發量
func (r *Round) Reader(rn int) *bytes.Pool {
	return &(r.readers[rn%r.options.Reader])
}

// 取Writer Pool，給定某數字以取餘數方式給其中一個，用於分散鎖競爭增加併發量
func (r *Round) Writer(rn int) *bytes.Pool {
	return &(r.writers[rn%r.options.Writer])
}

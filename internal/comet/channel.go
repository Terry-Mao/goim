package comet

import (
	"sync"

	"github.com/Terry-Mao/goim/api/comet/grpc"
	"github.com/Terry-Mao/goim/pkg/bufio"
)

// 用於推送消息給user，可以把這個識別user在聊天室內的地址
// 紀錄了當初連線至聊天室時所給的參數值
// 1. 身處在哪一個聊天室
// 2. user mid (user id)
// 3. user key
// 4. user operation
type Channel struct {
	// 該user進入的房間
	Room *Room

	// 讀寫異步的grpc.Proto緩型Pool
	CliProto Ring

	// 透過此管道處理Job service 送過來的資料
	signal chan *grpc.Proto

	// 用於寫操作的byte
	Writer bufio.Writer

	// 用於讀操作的byte
	Reader bufio.Reader

	// 雙向鏈結串列 rlink
	Next *Channel

	// 雙向鏈結串列 llink
	Prev *Channel

	// user id
	Mid int64

	// user在logic service的key
	Key string

	// user ip
	IP string

	// user只接收哪個房間id來的訊息
	watchOps map[int32]struct{}

	// 讀寫鎖
	mutex sync.RWMutex
}

// NewChannel new a channel.
func NewChannel(cli, svr int) *Channel {
	c := new(Channel)
	c.CliProto.Init(cli)

	// grpc接收資料的緩充量
	c.signal = make(chan *grpc.Proto, svr)
	c.watchOps = make(map[int32]struct{})
	return c
}

// 設置user 能接收到哪些房間來的訊息
func (c *Channel) Watch(accepts ...int32) {
	c.mutex.Lock()
	for _, op := range accepts {
		c.watchOps[op] = struct{}{}
	}
	c.mutex.Unlock()
}

// 移除user 不可在接收到某房間的訊息
func (c *Channel) UnWatch(accepts ...int32) {
	c.mutex.Lock()
	for _, op := range accepts {
		delete(c.watchOps, op)
	}
	c.mutex.Unlock()
}

// 判斷user能否接收來自某房間的訊息
func (c *Channel) NeedPush(op int32) bool {
	c.mutex.RLock()
	if _, ok := c.watchOps[op]; ok {
		c.mutex.RUnlock()
		return true
	}
	c.mutex.RUnlock()
	return false
}

// 針對某人推送訊息
func (c *Channel) Push(p *grpc.Proto) (err error) {
	// 當發送訊息速度大於消費速度則會阻塞
	// 使用select方式來避免這一塊但此時會有訊息丟失的風險存在
	// 可以提高signal buffer來避免但會耗費內存
	select {
	// 每個Channel會有自己signal接收處理的goroutine
	case c.signal <- p:
	default:
	}
	return
}

// 等待管道接收grpc資料
func (c *Channel) Ready() *grpc.Proto {
	return <-c.signal
}

// 接收到tcp資料傳遞給處理的goroutine
func (c *Channel) Signal() {
	c.signal <- grpc.ProtoReady
}

// 關閉連線flag
func (c *Channel) Close() {
	c.signal <- grpc.ProtoFinish
}

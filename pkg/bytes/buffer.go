package bytes

import (
	"sync"
)

// 一個表示單向鏈結串列表的Buffer，本身也負責管理某[]byte容量
type Buffer struct {
	// 管理的[]byte
	buf  []byte

	// 下一個鏈結的Buffer
	next *Buffer
}

func (b *Buffer) Bytes() []byte {
	return b.buf
}

// Pool 採用單向鏈結串列表來管理Buffer
// 每次free指向一個Buffer，而Buffer本身也紀錄下一個Buffer是誰，最後一個Buffer則會指向nil
// Pool內的Buffer只要生成完後就不會被GC，換句話說今天有100w+連線Pool就會
// 申請足夠應付100w+連線的內存，且當100w+的連線斷開後依然不會做GC
// 因為100w+斷線後下一秒可能會有100w+連線，所以不做GC
// 永遠保存最大內存做復用，以效能為優先
type Pool struct {
	// 鎖
	lock sync.Mutex

	// 當前Pool可供Get()拿走的Buffer
	free *Buffer

	// 每次跟os要的[]byte的總容量
	max  int

	// Buffer不夠時需多開多少個Buffer
	num  int

	// 每個Buffer的[]byte容量
	size int
}

// 建立一個Buffer Pool，給定num與size要Pool每次分配要管理多少個Buffer且byte容量為多少
func NewPool(num, size int) (p *Pool) {
	p = new(Pool)
	p.init(num, size)
	return
}

func (p *Pool) Init(num, size int) {
	p.init(num, size)
}

func (p *Pool) init(num, size int) {
	p.num = num
	p.size = size
	// 多少個Buffer，每個Buffer管理多少byte計算出總共需要申請多少byte容量
	p.max = num * size
	p.grow()
}

// 分配Buffer至Pool內
func (p *Pool) grow() {
	var (
		i   int
		b   *Buffer
		bs  []Buffer
		buf []byte
	)
	// 申請本次要分配給多個Buffer的總[]byte容量
	buf = make([]byte, p.max)

	// 初始化Pool申請的Buffer數量
	bs = make([]Buffer, p.num)

	// 將Pool free先指向第一個Buffer用於表示鏈結
	p.free = &bs[0]
	b = p.free

	// 根據Buffer數量來做單向鏈結串列表
	for i = 1; i < p.num; i++ {
		// 分配出每一個Buffer管理的byte容量
		b.buf = buf[(i-1)*p.size : i*p.size]

		// 第一個Buffer鏈結下一個Buffer
		b.next = &bs[i]
		b = b.next
	}

	// 最後一個鏈結也要分配byte容量
	b.buf = buf[(i-1)*p.size : i*p.size]

	// 最後一個鏈結因為已經到尾巴所以也不能鏈結下去
	b.next = nil
}

// 從Pool取出Buffer
// 每次取都會檢查Pool內管理的Buffer是否已經都被取走了
// 如果都被取走就跟os在申請Buffer
func (p *Pool) Get() (b *Buffer) {
	p.lock.Lock()
	if b = p.free; b == nil {
		p.grow()
		b = p.free
	}

	// 因為Pool採用單向鏈結串列管理所以每次都要指向下一個Buffer
	p.free = b.next
	p.lock.Unlock()
	return
}

// 用完的Buffer要歸還給Pool，假設Pool 有A -> B -> C，free指向A，Buffer間的單向鏈結串列還是A -> B -> C
// 當A與B已被取走時，則free指向C，Buffer間的單向鏈結串列還是A -> B -> C
// A歸還至Pool，則Buffer鏈結改為 A -> C與 B -> C，Pool free指向A
// B歸還至Pool，則Buffer鏈結改為 B -> A -> C，Pool free指向B
func (p *Pool) Put(b *Buffer) {
	p.lock.Lock()
	b.next = p.free
	p.free = b
	p.lock.Unlock()
}

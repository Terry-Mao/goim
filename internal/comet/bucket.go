package comet

import (
	"sync"
	"sync/atomic"

	"github.com/Terry-Mao/goim/api/comet/grpc"
	"github.com/Terry-Mao/goim/internal/comet/conf"
)

// 用於管理Room與Channel關於推送的邏輯
type Bucket struct {
	c *conf.Bucket

	// 讀寫鎖
	cLock sync.RWMutex

	// 當前管理的Channel以user sub key當作key
	chs map[string]*Channel

	// 當前管理的Room以Room id當作key
	rooms map[string]*Room

	// 一個Bucket會開多個goroutine，每個goroutine都有一把此chan針對房間做推送
	// 推送時會採用原子操作遞增並除於grpc.BroadcastRoomReq數量取餘數來決定使用哪一個
	routines []chan *grpc.BroadcastRoomReq

	// 用於決定由哪一個routines來做房間推送，此數字由atomic.AddUint64做原子操作遞增
	routinesNum uint64

	// 紀錄有哪些ip在此房間，key=ip value = 重覆ip數量
	ipCnts map[string]int32
}

// 初始化Bucket結構
func NewBucket(c *conf.Bucket) (b *Bucket) {
	b = new(Bucket)
	b.chs = make(map[string]*Channel, c.Channel)
	b.rooms = make(map[string]*Room, c.Room)
	b.ipCnts = make(map[string]int32)
	b.c = c

	// 設定該Bucket需要開幾個goroutine併發做房間訊息推送
	b.routines = make([]chan *grpc.BroadcastRoomReq, c.RoutineAmount)
	for i := uint64(0); i < c.RoutineAmount; i++ {
		c := make(chan *grpc.BroadcastRoomReq, c.RoutineSize)
		b.routines[i] = c
		go b.roomproc(c)
	}
	return
}

// 管理的人數
func (b *Bucket) ChannelCount() int {
	return len(b.chs)
}

// 管理的房間數量
func (b *Bucket) RoomCount() int {
	return len(b.rooms)
}

// 統計房間人數
func (b *Bucket) RoomsCount() (res map[string]int32) {
	var (
		roomID string
		room   *Room
	)
	b.cLock.RLock()
	res = make(map[string]int32)
	for roomID, room = range b.rooms {
		if room.Online > 0 {
			res[roomID] = room.Online
		}
	}
	b.cLock.RUnlock()
	return
}

// user更換房間
func (b *Bucket) ChangeRoom(nrid string, ch *Channel) (err error) {
	var (
		nroom *Room
		ok    bool
		oroom = ch.Room
	)
	// change to no room
	if nrid == "" {
		if oroom != nil && oroom.Del(ch) {
			b.DelRoom(oroom)
		}
		ch.Room = nil
		return
	}
	b.cLock.Lock()
	if nroom, ok = b.rooms[nrid]; !ok {
		nroom = NewRoom(nrid)
		b.rooms[nrid] = nroom
	}
	b.cLock.Unlock()
	if err = nroom.Put(ch); err != nil {
		return
	}
	ch.Room = nroom
	if oroom != nil && oroom.Del(ch) {
		b.DelRoom(oroom)
	}
	return
}

// 將user Channel 分配到某房間，總共會有三種結構互相關聯Bucket,Room,Channel
// 假設Room id = A , Channel key = B
// 1. Bucket put Channel(B)
// 2. Bucket put Room(A)
// 3. Channel(B) Room 對應到 Room(A)
// 4. Room(A) Channel put Channel(B)
// =============================================
// 		Bucket
//  		- []Room    			- []Channel
//			|						|
//			↓						↓
//		 Room(A) ←----------	 Channel(B) ←-|
//			- Channel		|-------- Room     |
//          |----------------------------------|
//
func (b *Bucket) Put(rid string, ch *Channel) (err error) {
	var (
		room *Room
		ok   bool
	)
	b.cLock.Lock()

	if dch := b.chs[ch.Key]; dch != nil {
		dch.Close()
	}
	b.chs[ch.Key] = ch
	if rid != "" {
		if room, ok = b.rooms[rid]; !ok {
			room = NewRoom(rid)
			b.rooms[rid] = room
		}
		ch.Room = room
	}
	b.ipCnts[ch.IP]++
	b.cLock.Unlock()
	if room != nil {
		err = room.Put(ch)
	}
	return
}

// 刪除某個user Channel
// 1. Bucket刪除[]Channel對應的Channel
// 2. Bucket刪除[]Room內對應的Channel
// 3. 如果Room沒人則Bucket刪除對應Room
func (b *Bucket) Del(dch *Channel) {
	var (
		ok   bool
		ch   *Channel
		room *Room
	)
	b.cLock.Lock()
	if ch, ok = b.chs[dch.Key]; ok {
		room = ch.Room
		if ch == dch {
			delete(b.chs, ch.Key)
		}
		if b.ipCnts[ch.IP] > 1 {
			b.ipCnts[ch.IP]--
		} else {
			delete(b.ipCnts, ch.IP)
		}
	}
	b.cLock.Unlock()
	if room != nil && room.Del(ch) {
		b.DelRoom(room)
	}
}

// 根據user key取對應Channel
func (b *Bucket) Channel(key string) (ch *Channel) {
	b.cLock.RLock()
	ch = b.chs[key]
	b.cLock.RUnlock()
	return
}

// 對Bucket內所有Channel且符合operation做訊息推送
func (b *Bucket) Broadcast(p *grpc.Proto, op int32) {
	var ch *Channel
	b.cLock.RLock()
	for _, ch = range b.chs {
		if !ch.NeedPush(op) {
			continue
		}
		_ = ch.Push(p)
	}
	b.cLock.RUnlock()
}

// 取得房間
func (b *Bucket) Room(rid string) (room *Room) {
	b.cLock.RLock()
	room = b.rooms[rid]
	b.cLock.RUnlock()
	return
}

// 刪除房間並將此房間內所有連線close
func (b *Bucket) DelRoom(room *Room) {
	b.cLock.Lock()
	delete(b.rooms, room.ID)
	b.cLock.Unlock()
	room.Close()
}

// logic service透過grpc推送給某個房間訊息
// Bucket本身會開多個goroutine做併發推送，與goroutine溝透透過chan
// 會使用原子鎖做遞增%Bucket開的goroutine數量來做選擇
func (b *Bucket) BroadcastRoom(arg *grpc.BroadcastRoomReq) {
	num := atomic.AddUint64(&b.routinesNum, 1) % b.c.RoutineAmount
	b.routines[num] <- arg
}

// Bucket內房間內所有人數大於1的房間id
func (b *Bucket) Rooms() (res map[string]struct{}) {
	var (
		roomID string
		room   *Room
	)
	res = make(map[string]struct{})
	b.cLock.RLock()
	for roomID, room = range b.rooms {
		if room.Online > 0 {
			res[roomID] = struct{}{}
		}
	}
	b.cLock.RUnlock()
	return
}

func (b *Bucket) IPCount() (res map[string]struct{}) {
	var (
		ip string
	)
	b.cLock.RLock()
	res = make(map[string]struct{}, len(b.ipCnts))
	for ip = range b.ipCnts {
		res[ip] = struct{}{}
	}
	b.cLock.RUnlock()
	return
}

func (b *Bucket) UpRoomsCount(roomCountMap map[string]int32) {
	var (
		roomID string
		room   *Room
	)
	b.cLock.RLock()
	for roomID, room = range b.rooms {
		room.AllOnline = roomCountMap[roomID]
	}
	b.cLock.RUnlock()
}

// 接收logic grpc client資料做某房間訊息推送
func (b *Bucket) roomproc(c chan *grpc.BroadcastRoomReq) {
	for {
		arg := <-c
		if room := b.Room(arg.RoomID); room != nil {
			room.Push(arg.Proto)
		}
	}
}

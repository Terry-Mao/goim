package comet

import (
	"sync"

	"github.com/Terry-Mao/goim/api/comet/grpc"
	"github.com/Terry-Mao/goim/internal/comet/errors"
)

// 房間結構，紀錄Channel採用雙向鏈結串列結構，房間內有A,B,C 三人結構如下，此三人都是Channel
// 			 	C <-> B <-> A         <==== 雙向鏈結串列
//  (next 指向)  ↑
// 		       Room
type Room struct {
	// 房間id
	ID string

	// 讀寫鎖
	rLock sync.RWMutex

	// 該房間管理的Channel，是一個雙向鏈結串列
	next *Channel

	// 房間人數是否為0
	drop bool

	// 房間總人數
	Online int32

	//
	AllOnline int32
}

// new 房間結構
func NewRoom(id string) (r *Room) {
	r = new(Room)
	r.ID = id
	r.drop = false
	r.next = nil
	r.Online = 0
	return
}

// 新增某個人進到該房間，假設房間現在沒有人
// 依序有A,B,C三個人要加入
// =======================================================================
// A加入: A next -> (指向) nil , room next -> A
// 目前結構:
//              A
//  (next 指向)  ↑
//     		   Room
// =======================================================================
// B加入: A prev -> B , B next -> A (做雙向鏈結串列) , B prev -> nil (如果之前有待過其他房間需把上下鏈結都清空才能完成切換房間)
//    room next -> B
// 目前結構:
//              B <-> A
//  (next 指向)  ↑
//     		   Room
// =======================================================================
// C加入: B prev -> C, C next -> A, C prev -> nil
//    room next -> C
// 目前結構:
//              C <-> B <-> A
//  (next 指向)  ↑
//     		   Room
//
func (r *Room) Put(ch *Channel) (err error) {
	r.rLock.Lock()
	// 房間
	if !r.drop {
		// 非第一個加入房間的人
		if r.next != nil {
			r.next.Prev = ch
		}

		ch.Next = r.next

		// 清掉prev鏈結清掉避免鏈結到上一個房間
		ch.Prev = nil
		r.next = ch
		r.Online++
	} else {
		err = errors.ErrRoomDroped
	}
	r.rLock.Unlock()
	return
}

// 某房間移除某個人，房間內有A,B,C
// 			 	C <-> B <-> A
//  (next 指向)  ↑
// 		       Room
// =======================================================================
// 只移除A: 把 B -> A 這條鏈結移除即可，找出A的prev(B)的next(A)，將此段變成nil
// 			 	C <-> B <- A
//  (next 指向)  ↑
// 		       Room
// =======================================================================
// 只移除B: 把 C -> B 與 B <- A 這兩條鏈結移除即可
//         1. 找出B的prev(C)的next(B)，將此段變成B的next(A)  ===> C -> A
//         2. 找出B的next(A)的prev(B)，將此段變成B的prev(C)  ===> C <- A
// 	   			-----------
//           	↓         ↑
//   Room  ->  	C <- B -> A
//     			↓         ↑
//     			-----------
// =======================================================================
// 只移除C: 把 C <- B 這條鏈，再把Room next指向C的next(B)
//        1. 找出C的next(B)的prev(C)，將此段變成B的prev(nil)
//        2. Room next 指向 C的next(B)
//
func (r *Room) Del(ch *Channel) bool {
	r.rLock.Lock()
	if ch.Next != nil {
		ch.Next.Prev = ch.Prev
	}
	if ch.Prev != nil {
		ch.Prev.Next = ch.Next
	} else {
		r.next = ch.Next
	}
	r.Online--
	r.drop = (r.Online == 0)
	r.rLock.Unlock()
	return r.drop
}

// 單一房間所有人的訊息推送
func (r *Room) Push(p *grpc.Proto) {
	r.rLock.RLock()
	// Channel採用雙向鏈結串列，所以用for往next找直到nil
	for ch := r.next; ch != nil; ch = ch.Next {
		_ = ch.Push(p)
	}
	r.rLock.RUnlock()
}

// 關閉房間內所有的Client連線
func (r *Room) Close() {
	r.rLock.RLock()
	// Channel採用雙向鏈結串列，所以用for往next找直到nil
	for ch := r.next; ch != nil; ch = ch.Next {
		ch.Close()
	}
	r.rLock.RUnlock()
}

func (r *Room) OnlineNum() int32 {
	if r.AllOnline > 0 {
		return r.AllOnline
	}
	return r.Online
}

package logic

import (
	"context"
	"sort"
	"strings"

	"github.com/Terry-Mao/goim/internal/logic/model"
)

var (
	_emptyTops = make([]*model.Top, 0)
)

// 依照房間總人數取前幾名的房間
func (l *Logic) OnlineTop(c context.Context, typ string, n int) (tops []*model.Top, err error) {
	// 由於room id的規則是type://id
	// 所以要針對這個規則做字串分割以type來分
	// key:就是room id
	// cnt是房間在線人數
	for key, cnt := range l.roomCount {
		if strings.HasPrefix(key, typ) {
			// type://id切成type與id
			_, roomID, err := model.DecodeRoomKey(key)
			if err != nil {
				continue
			}
			top := &model.Top{
				RoomID: roomID,
				Count:  cnt,
			}
			tops = append(tops, top)
		}
	}

	// 房間依照在線人數做排序
	sort.Slice(tops, func(i, j int) bool {
		return tops[i].Count > tops[j].Count
	})

	// 只取前幾個房間
	if len(tops) > n {
		tops = tops[:n]
	}
	if len(tops) == 0 {
		tops = _emptyTops
	}
	return
}

// 根據房間type與room id取房間在線人數
func (l *Logic) OnlineRoom(c context.Context, typ string, rooms []string) (res map[string]int32, err error) {
	res = make(map[string]int32, len(rooms))
	for _, room := range rooms {
		res[room] = l.roomCount[model.EncodeRoomKey(typ, room)]
	}
	return
}

//
func (l *Logic) OnlineTotal(c context.Context) (int64, int64) {
	return l.totalIPs, l.totalConns
}

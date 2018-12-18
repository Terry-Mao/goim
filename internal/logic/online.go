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

// OnlineTop get the top online.
func (l *Logic) OnlineTop(c context.Context, typ string, n int) (tops []*model.Top, err error) {
	for roomKey, cnt := range l.roomCount {
		if strings.HasPrefix(roomKey, typ) {
			_, roomID, err := model.DecodeRoomKey(roomKey)
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
	sort.Slice(tops, func(i, j int) bool {
		return tops[i].Count > tops[j].Count
	})
	if len(tops) > n {
		tops = tops[:n]
	}
	if len(tops) == 0 {
		tops = _emptyTops
	}
	return
}

// OnlineRoom get rooms online.
func (l *Logic) OnlineRoom(c context.Context, rooms []string) (res map[string]int32, err error) {
	res = make(map[string]int32, len(rooms))
	for _, roomID := range rooms {
		res[roomID] = l.roomCount[roomID]
	}
	return
}

// OnlineTotal get all online.
func (l *Logic) OnlineTotal(c context.Context) (int64, int64) {
	return l.totalIPs, l.totalConns
}

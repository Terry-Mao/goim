package main

import (
	"goim/libs/net/xrpc"
	"time"
)

const (
	syncCountDelay = 1 * time.Second
)

var (
	RoomCountMap   = make(map[int32]int32) // roomid:count
	ServerCountMap = make(map[int32]int32) // server:count
)

func MergeCount() {
	var (
		c                     *xrpc.Clients
		err                   error
		roomId, server, count int32
		counter               map[int32]int32
		roomCount             = make(map[int32]int32)
		serverCount           = make(map[int32]int32)
	)
	// all comet nodes
	for _, c = range routerServiceMap {
		if c != nil {
			if counter, err = allRoomCount(c); err != nil {
				continue
			}
			for roomId, count = range counter {
				roomCount[roomId] += count
			}
			if counter, err = allServerCount(c); err != nil {
				continue
			}
			for server, count = range counter {
				serverCount[server] += count
			}
		}
	}
	RoomCountMap = roomCount
	ServerCountMap = serverCount
}

/*
func RoomCount(roomId int32) (count int32) {
	count = RoomCountMap[roomId]
	return
}
*/

func SyncCount() {
	for {
		MergeCount()
		time.Sleep(syncCountDelay)
	}
}

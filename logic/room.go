package main

import (
	"github.com/Terry-Mao/protorpc"
	"time"
)

const (
	syncRoomCountDelay = 1 * time.Second
)

var (
	RoomCountMap = make(map[int32]int32) // roomid:count
)

func MergeRoomCount() {
	var (
		c             **protorpc.Client
		err           error
		roomId, count int32
		counter       map[int32]int32
		roomCount     = make(map[int32]int32)
	)
	// all comet nodes
	for _, c = range routerServiceMap {
		if *c != nil {
			if counter, err = allRoomCount(*c); err != nil {
				continue
			}
			for roomId, count = range counter {
				roomCount[roomId] += count
			}
		}
	}
	RoomCountMap = roomCount
}

func RoomCount(roomId int32) (count int32) {
	count = RoomCountMap[roomId]
	return
}

func SyncRoomCount() {
	for {
		MergeRoomCount()
		time.Sleep(syncRoomCountDelay)
	}
}

package main

import (
	"time"
)

const (
	syncRoomServersDelay = 1 * time.Second
)

var (
	RoomServersMap = make(map[int32]map[int32]struct{}) // roomid:servers
)

func MergeRoomServers() {
	var (
		ok          bool
		roomId      int32
		rooms       map[int32]bool
		servers     map[int32]struct{}
		roomServers = make(map[int32]map[int32]struct{})
	)
	// all comet nodes
	for serverId, c := range cometServiceMap {
		if *c != nil {
			if rooms = roomsComet(*c); rooms != nil {
				// merge room's servers
				for roomId, _ = range rooms {
					if servers, ok = roomServers[roomId]; !ok {
						servers = make(map[int32]struct{})
						roomServers[roomId] = servers
					}
					servers[serverId] = struct{}{}
				}
			}
		}
	}
}

func SyncRoomServers() {
	for {
		MergeRoomServers()
		time.Sleep(syncRoomServersDelay)
	}
}

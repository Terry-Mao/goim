package main

import (
	"net/rpc"
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
		c           **rpc.Client
		ok          bool
		roomId      int32
		serverId    int32
		roomIds     []int32
		servers     map[int32]struct{}
		roomServers = make(map[int32]map[int32]struct{})
	)
	// all comet nodes
	for serverId, c = range cometServiceMap {
		if *c != nil {
			if roomIds = roomsComet(*c); roomIds != nil {
				// merge room's servers
				for _, roomId = range roomIds {
					if servers, ok = roomServers[roomId]; !ok {
						servers = make(map[int32]struct{})
						roomServers[roomId] = servers
					}
					servers[serverId] = struct{}{}
				}
			}
		}
	}
	RoomServersMap = roomServers
}

func SyncRoomServers() {
	for {
		MergeRoomServers()
		time.Sleep(syncRoomServersDelay)
	}
}

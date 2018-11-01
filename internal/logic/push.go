package service

import (
	"context"

	log "github.com/golang/glog"
)

// PushKeys push a message by keys.
func (s *Service) PushKeys(c context.Context, op int32, keys []string, msg string) (err error) {
	servers, err := s.dao.ServersByKeys(c, keys)
	if err != nil {
		return
	}
	pushKeys := make(map[string][]string)
	for i, key := range keys {
		server := servers[i]
		if server != "" && key != "" {
			pushKeys[server] = append(pushKeys[server], key)
		}
	}
	for server := range pushKeys {
		if err = s.dao.PushMsg(c, op, server, msg, pushKeys[server]); err != nil {
			return
		}
	}
	return
}

// PushMids push a message by mid.
func (s *Service) PushMids(c context.Context, op int32, mids []int64, msg string) (err error) {
	keyServers, _, err := s.dao.KeysByMids(c, mids)
	if err != nil {
		return
	}
	keys := make(map[string][]string)
	for key, server := range keyServers {
		if key != "" && server != "" {
			keys[server] = append(keys[server], key)
		} else {
			log.Warningf("push key:%s server:%s is empty", key, server)
		}
	}
	for server, keys := range keys {
		if err = s.dao.PushMsg(c, op, server, msg, keys); err != nil {
			return
		}
	}
	return
}

// PushRoom push a message by room.
func (s *Service) PushRoom(c context.Context, op int32, room, msg string) (err error) {
	if err = s.dao.BroadcastRoomMsg(c, op, room, msg); err != nil {
		return
	}
	return
}

// PushAll push a message to all.
func (s *Service) PushAll(c context.Context, op, speed int32, msg, platform string) (err error) {
	if err = s.dao.BroadcastMsg(c, op, speed, msg, platform); err != nil {
		return
	}
	return
}

package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Terry-Mao/goim/internal/logic/model"
	xstr "github.com/Terry-Mao/goim/pkg/strings"
	log "github.com/golang/glog"
)

// Connect connect a conn.
func (s *Service) Connect(c context.Context, server, serverKey, cookie string, token []byte) (mid int64, key, roomID string, paltform string, accepts []int32, err error) {
	// TODO test example: mid|key|roomid|platform|accepts
	params := strings.Split(string(token), "|")
	if len(params) != 5 {
		err = fmt.Errorf("invalid token:%s", token)
		return
	}
	if mid, err = strconv.ParseInt(params[0], 10, 64); err != nil {
		return
	}
	key = params[1]
	roomID = params[2]
	paltform = params[3]
	if accepts, err = xstr.SplitInt32s(params[4], ","); err != nil {
		return
	}
	log.Infof("conn connected key:%s server:%s mid:%d token:%s", key, server, mid, token)
	return
}

// Disconnect disconnect a conn.
func (s *Service) Disconnect(c context.Context, mid int64, key, server string) (has bool, err error) {
	if has, err = s.dao.DelMapping(c, mid, key, server); err != nil {
		log.Errorf("s.dao.DelMapping(%d,%s) error(%v)", mid, key, server)
		return
	}
	log.Infof("conn disconnected key:%s server:%s mid:%d", key, server, mid)
	return
}

// Heartbeat heartbeat a conn.
func (s *Service) Heartbeat(c context.Context, mid int64, key, server string) (err error) {
	has, err := s.dao.ExpireMapping(c, mid, key)
	if err != nil {
		log.Errorf("s.dao.ExpireMapping(%d,%s,%s) error(%v)", mid, key, server, err)
		return
	}
	if !has {
		if err = s.dao.AddMapping(c, mid, key, server); err != nil {
			log.Errorf("s.dao.AddMapping(%d,%s,%s) error(%v)", mid, key, server, err)
			return
		}
	}
	log.Infof("conn heartbeat key:%s server:%s mid:%d", key, server, mid)
	return
}

// RenewServer renew a server info.
func (s *Service) RenewServer(c context.Context, server string, ipAddrs []string, ipCount, connCount int32, shutdown bool) (err error) {
	if shutdown {
		s.dao.DelServerInfo(c, server)
		return
	}
	serverInfo := &model.ServerInfo{
		Server:    server,
		IPAddrs:   ipAddrs,
		IPCount:   ipCount,
		ConnCount: connCount,
		Updated:   time.Now().Unix(),
	}
	if err = s.dao.AddServerInfo(c, server, serverInfo); err != nil {
		return
	}
	return
}

// RenewOnline renew a server online.
func (s *Service) RenewOnline(c context.Context, server string, roomCount map[string]int32) (allRoomCount map[string]int32, err error) {
	online := &model.Online{
		Server:    server,
		RoomCount: roomCount,
		Updated:   time.Now().Unix(),
	}
	if err = s.dao.AddServerOnline(context.Background(), server, online); err != nil {
		return
	}
	return s.roomCount, nil
}

// Receive receive a message.
func (s *Service) Receive(c context.Context, mid int64) (err error) {
	// TODO upstream message
	log.Infof("conn receive a message mid:%d", mid)
	return
}

package main

import (
	"encoding/json"
	"goim/libs/define"
	"goim/libs/proto"

	"github.com/nsqio/go-nsq"
)

const (
	NsqPushsTopic = "NsqPushsTopic"
)

var (
	nsqProducer *nsq.Producer
)

func InitNsq(addr string) (err error) {
	nsqProducer, err = nsq.NewProducer(addr, nsq.NewConfig())
	return
}

func mpushNsq(serverId int32, keys []string, msg []byte) (err error) {
	var (
		vBytes []byte
		v      = &proto.NsqMsg{OP: define.NSQ_MESSAGE_MULTI, ServerId: serverId, SubKeys: keys, Msg: msg}
	)
	if vBytes, err = json.Marshal(v); err != nil {
		return
	}
	err = nsqProducer.Publish(NsqPushsTopic, vBytes)
	return
}

func broadcastNsq(msg []byte) (err error) {
	var (
		vBytes []byte
		v      = &proto.NsqMsg{OP: define.NSQ_MESSAGE_BROADCAST, Msg: msg}
	)
	if vBytes, err = json.Marshal(v); err != nil {
		return
	}
	err = nsqProducer.Publish(NsqPushsTopic, vBytes)
	return
}

func broadcastRoomNsq(rid int32, msg []byte, ensure bool) (err error) {
	var (
		vBytes   []byte
		v        = &proto.NsqMsg{OP: define.NSQ_MESSAGE_BROADCAST_ROOM, RoomId: rid, Msg: msg, Ensure: ensure}
	)
	if vBytes, err = json.Marshal(v); err != nil {
		return
	}
	err = nsqProducer.Publish(NsqPushsTopic, vBytes)
	return
}

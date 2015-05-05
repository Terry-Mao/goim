package main

import (
	log "code.google.com/p/log4go"
	"strconv"
	"sync/atomic"
	"time"
)

const (
	OP_HANDSHARE        = uint32(0)
	OP_HANDSHARE_REPLY  = uint32(1)
	OP_HEARTBEAT        = uint32(2)
	OP_HEARTBEAT_REPLY  = uint32(3)
	OP_SEND_SMS         = uint32(4)
	OP_SEND_SMS_REPLY   = uint32(5)
	OP_DISCONNECT_REPLY = uint32(6)

	// for test
	OP_TEST       = uint32(254)
	OP_TEST_REPLY = uint32(255)
)

var (
	testKey = int64(0)
)

type IMOperator struct {
}

func (operator *IMOperator) Operate(proto *Proto) error {
	if proto.Operation == OP_HEARTBEAT {
		proto.Body = nil
		proto.Operation = OP_HEARTBEAT_REPLY
		log.Info("heartbeat proto: %v", proto)
		return nil
	} else if proto.Operation == OP_SEND_SMS {
		// call suntao's api
		// proto.Body = nil
		proto.Operation = OP_SEND_SMS_REPLY
		log.Info("send sms proto: %v", proto)
		return nil
	} else if proto.Operation == OP_TEST {
		log.Debug("test operation: %s", proto.Body)
		proto.Operation = OP_TEST_REPLY
		proto.Body = []byte("reply test")
		return nil
	}
	return nil
}

func (operator *IMOperator) Connect(body []byte) (string, time.Duration, error) {
	// TODO call register router
	atomic.AddInt64(&testKey, 1)
	return "Terry-Mao" + strconv.FormatInt(testKey, 10), 30 * time.Second, nil
}

func (operator *IMOperator) Disconnect(string) error {
	return nil
}

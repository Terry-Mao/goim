package main

import (
	log "code.google.com/p/log4go"
	"time"
)

const (
	// handshake
	OP_HANDSHARE       = int32(0)
	OP_HANDSHARE_REPLY = int32(1)
	// heartbeat
	OP_HEARTBEAT       = int32(2)
	OP_HEARTBEAT_REPLY = int32(3)
	// send text messgae
	OP_SEND_SMS       = int32(4)
	OP_SEND_SMS_REPLY = int32(5)
	// kick user
	OP_DISCONNECT_REPLY = int32(6)
	// auth user
	OP_AUTH       = int32(7)
	OP_AUTH_REPLY = int32(8)

	// for test
	OP_TEST       = int32(254)
	OP_TEST_REPLY = int32(255)
)

type Operator interface {
	// Operate process the common operation such as send message etc.
	Operate(*Proto) error
	// Connect used for auth user and return a sub key & hearbeat.
	Connect(*Proto) (string, time.Duration, error)
	// Disconnect used for revoke the subkey.
	Disconnect(string) error
}

type IMOperator struct {
}

func (operator *IMOperator) Operate(proto *Proto) error {
	if proto.Operation == OP_SEND_SMS {
		// call suntao's api
		// proto.Body = nil

		proto.Operation = OP_SEND_SMS_REPLY
		log.Info("send sms proto: %v", proto)
		return nil
	} else if proto.Operation == OP_TEST {
		// TODO remove
		log.Debug("test operation: %s", proto.Body)
		proto.Operation = OP_TEST_REPLY
		proto.Body = []byte("reply test")
		return nil
	} else {
		return ErrOperation
	}
	return nil
}

func (operator *IMOperator) Connect(proto *Proto) (subKey string, heartbeat time.Duration, err error) {
	// TODO call register router
	// for test
	subKey = string(proto.Body)
	heartbeat = 30 * time.Second
	return
}

func (operator *IMOperator) Disconnect(subKey string) error {
	return nil
}

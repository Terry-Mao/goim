package main

import (
	log "code.google.com/p/log4go"
	"time"
)

const (
	// handshake
	OP_HANDSHAKE       = int32(0)
	OP_HANDSHAKE_REPLY = int32(1)
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
	// handshake with sid
	OP_HANDSHAKE_SID       = int32(9)
	OP_HANDSHAKE_SID_REPLY = int32(10)

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

type DefaultOperator struct {
}

func (operator *DefaultOperator) Operate(p *Proto) error {
	var (
		body []byte
	)
	if p.Operation == OP_SEND_SMS {
		// call suntao's api
		// p.Body = nil
		p.Operation = OP_SEND_SMS_REPLY
		log.Info("send sms proto: %v", p)
	} else if p.Operation == OP_TEST {
		log.Debug("test operation: %s", body)
		p.Operation = OP_TEST_REPLY
		p.Body = []byte("{\"test\":\"come on\"}")
	} else {
		return ErrOperation
	}
	return nil
}

func (operator *DefaultOperator) Connect(p *Proto) (key string, heartbeat time.Duration, err error) {
	key, heartbeat, err = connect(p)
	return
}

func (operator *DefaultOperator) Disconnect(key string) (err error) {
	var has bool
	if has, err = disconnect(key); err != nil {
		return
	}
	if has {
		log.Warn("disconnect key: \"%s\" not exists", key)
	}
	return
}

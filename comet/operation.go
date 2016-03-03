package main

import (
	"goim/libs/define"
	"goim/libs/proto"
	"time"

	log "github.com/thinkboy/log4go"
)

type Operator interface {
	// Operate process the common operation such as send message etc.
	Operate(*proto.Proto) error
	// Connect used for auth user and return a subkey, roomid, hearbeat.
	Connect(*proto.Proto) (string, int32, time.Duration, error)
	// Disconnect used for revoke the subkey.
	Disconnect(string, int32) error
}

type DefaultOperator struct {
}

func (operator *DefaultOperator) Operate(p *proto.Proto) error {
	var (
		body []byte
	)
	if p.Operation == define.OP_SEND_SMS {
		// call suntao's api
		// p.Body = nil
		p.Operation = define.OP_SEND_SMS_REPLY
		log.Info("send sms proto: %v", p)
	} else if p.Operation == define.OP_TEST {
		log.Debug("test operation: %s", body)
		p.Operation = define.OP_TEST_REPLY
		p.Body = []byte("{\"test\":\"come on\"}")
	} else {
		return ErrOperation
	}
	return nil
}

func (operator *DefaultOperator) Connect(p *proto.Proto) (key string, rid int32, heartbeat time.Duration, err error) {
	key, rid, heartbeat, err = connect(p)
	return
}

func (operator *DefaultOperator) Disconnect(key string, rid int32) (err error) {
	var has bool
	if has, err = disconnect(key, rid); err != nil {
		return
	}
	if !has {
		log.Warn("disconnect key: \"%s\" not exists", key)
	}
	return
}

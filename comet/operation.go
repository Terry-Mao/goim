package main

import (
	log "code.google.com/p/log4go"
	"github.com/Terry-Mao/goim/define"
	"time"
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

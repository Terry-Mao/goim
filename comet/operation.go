package main

import (
	log "code.google.com/p/log4go"
)

const (
	OP_HANDSHARE       = uint32(0)
	OP_HANDSHARE_REPLY = uint32(1)
	OP_HEARTBEAT       = uint32(2)
	OP_HEARTBEAT_REPLY = uint32(3)
	OP_SEND_SMS        = uint32(4)
	OP_SEND_SMS_REPLY  = uint32(5)
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
	}
	return nil
}

func (operator *IMOperator) Connect(body []byte) (string, error) {
	// TODO call register router
	return "Terry-Mao", nil
}

func (operator *IMOperator) Disconnect(string) error {
	return nil
}

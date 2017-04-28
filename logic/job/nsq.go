package main

import (
	"github.com/nsqio/go-nsq"
)

const (
	NsqConsumeTopic = "NsqPushsTopic"
	NsqConsumeChannel = "NsqPushsChannel"
)

type ConsumerT struct{}

func (*ConsumerT) HandleMessage(msg *nsq.Message) error {
	push(msg.Body)
	return nil
}

func InitNsq() error {
	c, err := nsq.NewConsumer(NsqConsumeTopic, NsqConsumeChannel, nsq.NewConfig())
	if err != nil {
		return err
	}
	c.AddHandler(&ConsumerT{})
	if err := c.ConnectToNSQD(Conf.NsqAddr); err != nil {
		return err
	}

	return nil
}

package main

import (
	"encoding/json"

	log "code.google.com/p/log4go"
	"github.com/Shopify/sarama"
	"github.com/thinkboy/goim/define"
)

const (
	KafkaPushsTopic = "KafkaPushsTopic"
)

var (
	producer sarama.SyncProducer
)

func InitKafka(kafkaAddrs []string) (err error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	producer, err = sarama.NewSyncProducer(kafkaAddrs, config)
	return
}

func multiPushTokafka(userIds []int64, msg []byte) (err error) {
	var (
		vBytes []byte
		v      = &define.KafkaPushsMsg{UserIds: userIds, Msg: msg}
	)
	if vBytes, err = json.Marshal(v); err != nil {
		return
	}
	message := &sarama.ProducerMessage{Topic: KafkaPushsTopic, Key: sarama.StringEncoder(define.KAFKA_MESSAGE_MULTI), Value: sarama.ByteEncoder(vBytes)}
	if _, _, err = producer.SendMessage(message); err != nil {
		return
	}
	log.Debug("produce msg ok, userids:%v msg:%s", userIds, msg)
	return
}

func broadcastTokafka(msg []byte) (err error) {
	message := &sarama.ProducerMessage{Topic: KafkaPushsTopic, Key: sarama.StringEncoder(define.KAFKA_MESSAGE_BROADCAST), Value: sarama.ByteEncoder(msg)}
	if _, _, err = producer.SendMessage(message); err != nil {
		return
	}
	log.Debug("produce msg ok, broadcast msg:%s", msg)
	return
}

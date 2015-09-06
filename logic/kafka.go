package main

import (
	"github.com/Shopify/sarama"
	"github.com/Terry-Mao/goim/define"
	lproto "github.com/Terry-Mao/goim/proto/logic"
	"github.com/gogo/protobuf/proto"
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

func mpushKafka(server int32, keys []string, msg []byte) (err error) {
	var (
		vBytes []byte
		v      = &lproto.PushsMsg{Server: server, SubKeys: keys, Msg: msg}
	)
	if vBytes, err = proto.Marshal(v); err != nil {
		return
	}
	message := &sarama.ProducerMessage{Topic: KafkaPushsTopic, Key: sarama.StringEncoder(define.KAFKA_MESSAGE_MULTI), Value: sarama.ByteEncoder(vBytes)}
	if _, _, err = producer.SendMessage(message); err != nil {
		return
	}
	return
}

func broadcastKafka(msg []byte) (err error) {
	message := &sarama.ProducerMessage{Topic: KafkaPushsTopic, Key: sarama.StringEncoder(define.KAFKA_MESSAGE_BROADCAST), Value: sarama.ByteEncoder(msg)}
	if _, _, err = producer.SendMessage(message); err != nil {
		return
	}
	return
}

func broadcastRoomKafka(ridStr string, msg []byte) (err error) {
	message := &sarama.ProducerMessage{Topic: KafkaPushsTopic, Key: sarama.StringEncoder(ridStr), Value: sarama.ByteEncoder(msg)}
	if _, _, err = producer.SendMessage(message); err != nil {
		return
	}
	return
}

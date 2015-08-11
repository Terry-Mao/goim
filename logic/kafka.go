package main

import (
	log "code.google.com/p/log4go"
	"encoding/binary"
	"encoding/json"
	"github.com/Shopify/sarama"
)

const (
	// KAFKA_TOPIC_PUSH  = "kafka_topic_push"
	KafkaPushsTopic = "KafkaPushsTopic"
)

var (
	producer sarama.SyncProducer
)

type KafkaPushs struct {
	Subkeys []string `json:"subkeys"`
	Msg     []byte   `json:"msg"`
}

func InitKafka(kafkaAddrs []string) (err error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Partitioner = sarama.NewHashPartitioner
	producer, err = sarama.NewSyncProducer(kafkaAddrs, config)
	return
}

/*
func pushTokafka(userID int64, value []byte) (err error) {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(userID))
	message := &sarama.ProducerMessage{Topic: KAFKA_TOPIC_PUSH, Key: sarama.ByteEncoder(b), Value: sarama.ByteEncoder(value)}
	if _, _, err = producer.SendMessage(message); err != nil {
		return
	}
	log.Debug("produce msg ok, key:%s", strconv.FormatInt(userID, 10))
	return
}
*/

func pushsTokafka(serverId int32, subkeys []string, msg []byte) (err error) {
	var (
		vBytes []byte
		v      = &KafkaPushs{Subkeys: subkeys, Msg: msg}
		b      = make([]byte, 8)
	)
	if vBytes, err = json.Marshal(v); err != nil {
		return
	}
	binary.BigEndian.PutUint64(b, uint64(serverId))
	message := &sarama.ProducerMessage{Topic: KafkaPushsTopic, Key: sarama.ByteEncoder(b), Value: sarama.ByteEncoder(vBytes)}
	if _, _, err = producer.SendMessage(message); err != nil {
		return
	}
	log.Debug("produce msg ok, serverId:%d subkeys:%v", serverId, subkeys)
	return
}

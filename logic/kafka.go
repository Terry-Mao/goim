package main

import (
	log "code.google.com/p/log4go"
	"encoding/binary"
	"encoding/json"
	"github.com/Shopify/sarama"
	"strconv"
)

const (
	KAFKA_TOPIC_PUSH  = "kafka_topic_push"
	KAFKA_TOPIC_PUSHS = "kafka_topic_pushs"
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
	if err != nil {
		return err
	}
	return nil
}

//TODO:考虑是否使用异步
func pushTokafka(userID int64, value []byte) (err error) {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(userID))
	message := &sarama.ProducerMessage{Topic: KAFKA_TOPIC_PUSH, Key: sarama.ByteEncoder(b), Value: sarama.ByteEncoder(value)}
	_, _, err = producer.SendMessage(message)
	if err != nil {
		return
	}
	log.Info("produce msg ok, key:%s", strconv.FormatInt(userID, 10))
	return
}

//TODO:考虑是否使用异步
func pushsTokafka(serverId int32, subkeys []string, msg []byte) (err error) {
	var (
		vBytes []byte
		v      = &KafkaPushs{Subkeys: subkeys, Msg: msg}
	)
	vBytes, err = json.Marshal(v)
	if err != nil {
		return
	}
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(serverId))
	message := &sarama.ProducerMessage{Topic: KAFKA_TOPIC_PUSHS, Key: sarama.ByteEncoder(b), Value: sarama.ByteEncoder(vBytes)}
	_, _, err = producer.SendMessage(message)
	if err != nil {
		return
	}
	log.Info("produce msg ok, serverId:%d subkeys:%v", serverId, subkeys)
	return
}

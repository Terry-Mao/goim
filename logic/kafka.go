package main

import (
	log "code.google.com/p/log4go"
	"encoding/binary"
	"github.com/Shopify/sarama"
	"strconv"
)

const (
	KAFKA_TOPIC_PUSH = "kafka_topic_push"
)

var (
	producer sarama.SyncProducer
)

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

//TODO:考虑使用异步
func pushTokafka(userID int64, value []byte) (err error) {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(userID))
	message := &sarama.ProducerMessage{Topic: KAFKA_TOPIC_PUSH, Key: sarama.ByteEncoder(b), Value: sarama.ByteEncoder(value)}
	_, _, err = producer.SendMessage(message)
	if err != nil {
		log.Error("producer.SendMessage(message)  key:%s error(%v)", strconv.FormatInt(userID, 10), err)
		return
	}
	log.Info("produce msg ok, key:%s", strconv.FormatInt(userID, 10))
	return
}

package main

import (
	llog "log"
	"os"
	"time"

	"github.com/Shopify/sarama"
	log "github.com/thinkboy/log4go"
	"github.com/wvanbergen/kafka/consumergroup"
)

const (
	KAFKA_GROUP_NAME                   = "kafka_topic_push_group"
	OFFSETS_PROCESSING_TIMEOUT_SECONDS = 10 * time.Second
	OFFSETS_COMMIT_INTERVAL            = 10 * time.Second
)

func InitKafka() error {
	log.Info("start topic:%s consumer", Conf.KafkaTopic)
	log.Info("consumer group name:%s", KAFKA_GROUP_NAME)
	sarama.Logger = llog.New(os.Stdout, "[Sarama] ", llog.LstdFlags)
	config := consumergroup.NewConfig()
	config.Offsets.Initial = sarama.OffsetNewest
	config.Offsets.ProcessingTimeout = OFFSETS_PROCESSING_TIMEOUT_SECONDS
	config.Offsets.CommitInterval = OFFSETS_COMMIT_INTERVAL
	config.Zookeeper.Chroot = Conf.ZKRoot
	kafkaTopics := []string{Conf.KafkaTopic}
	cg, err := consumergroup.JoinConsumerGroup(KAFKA_GROUP_NAME, kafkaTopics, Conf.ZKAddrs, config)
	if err != nil {
		return err
	}
	go func() {
		for err := range cg.Errors() {
			log.Error("consumer error(%v)", err)
		}
	}()
	go func() {
		for msg := range cg.Messages() {
			log.Info("deal with topic:%s, partitionId:%d, Offset:%d, Key:%s msg:%s", msg.Topic, msg.Partition, msg.Offset, msg.Key, msg.Value)
			push(msg.Value)
			cg.CommitUpto(msg)
		}
	}()
	return nil
}

package main

import (
	log "code.google.com/p/log4go"
	"github.com/Shopify/sarama"
	"github.com/Terry-Mao/goim/define"
	lproto "github.com/Terry-Mao/goim/proto/logic"
	"github.com/gogo/protobuf/proto"
	"github.com/wvanbergen/kafka/consumergroup"
	"time"
)

const (
	KAFKA_GROUP_NAME                   = "kafka_topic_push_group"
	OFFSETS_PROCESSING_TIMEOUT_SECONDS = 10 * time.Second
	OFFSETS_COMMIT_INTERVAL            = 10 * time.Second
)

func InitKafka() error {
	log.Info("start topic:%s consumer", Conf.KafkaTopic)
	log.Info("consumer group name:%s", KAFKA_GROUP_NAME)
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
			push(string(msg.Key), msg.Value)
			cg.CommitUpto(msg)
		}
	}()
	return nil
}

func push(op string, msg []byte) (err error) {
	if op == define.KAFKA_MESSAGE_MULTI {
		m := &lproto.PushsMsg{}
		if err = proto.Unmarshal(msg, m); err != nil {
			log.Error("proto.Unmarshal(%s) serverId:%d error(%s)", msg, err)
			return
		}
		mpush(m.Server, m.SubKeys, m.Msg)
	} else if op == define.KAFKA_MESSAGE_BROADCAST {
		broadcast(msg)
	} else {
		log.Error("unknown message type:%s", op)
	}
	return
}

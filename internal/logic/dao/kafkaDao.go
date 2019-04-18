package dao

import (
	"github.com/Shopify/sarama"
	kafka "gopkg.in/Shopify/sarama.v1"

	"github.com/Terry-Mao/goim/internal/logic/conf"
)

// Dao dao.
type kafkaDao struct {
	c    *conf.Config
	push kafka.SyncProducer
}

// New new a dao and return.
func NewKafka(c *conf.Config) *kafkaDao {
	d := &kafkaDao{
		c:    c,
		push: newKafkaPub(c.Kafka),
	}
	return d
}

// PublishMessage  push message to kafka
func (d *kafkaDao) PublishMessage(topic, ackInbox string, key string, value []byte) error {

	m := &kafka.ProducerMessage{
		Key:   sarama.StringEncoder(key),
		Topic: d.c.Kafka.Topic,
		Value: sarama.ByteEncoder(value),
	}
	_, _, err := d.push.SendMessage(m)

	return err
}

// Close close the resource.
func (d *kafkaDao) Close() error {
	return d.push.Close()
}

func newKafkaPub(c *conf.Kafka) kafka.SyncProducer {
	var err error
	kc := kafka.NewConfig()
	kc.Producer.RequiredAcks = kafka.WaitForAll // Wait for all in-sync replicas to ack the message
	kc.Producer.Retry.Max = 10                  // Retry up to 10 times to produce the message
	kc.Producer.Return.Successes = true
	pub, err := kafka.NewSyncProducer(c.Brokers, kc)
	if err != nil {
		panic(err)
	}
	return pub
}

package dao

type PushMsg interface {
	PublishMessage(topic, ackInbox string, key string, msg []byte) error
	Close() error
}

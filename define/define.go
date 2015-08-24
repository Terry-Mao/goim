package define

// Kafka message type Commands
const (
	KAFKA_MESSAGE_MULTI     = "multiple"  //multi-userid push
	KAFKA_MESSAGE_BROADCAST = "broadcast" //broadcast push
)

type KafkaPushsMsg struct {
	UserIds []int64 `json:"userids"`
	Msg     []byte  `json:"msg"`
}

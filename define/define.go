package define

// Kafka message type Commands
const (
	KAFKA_MESSAGE_MULTI     = "multiple"  //multi-userid push
	KAFKA_MESSAGE_BROADCAST = "broadcast" //broadcast push
)

type KafkaPushsMsg struct {
	CometIds []int32    `cometid`
	Subkeys  [][]string `json:"subkeys"`
	Msg      []byte     `json:"msg"`
}

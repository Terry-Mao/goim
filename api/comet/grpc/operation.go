package grpc

const (
	//
	OpHandshake = int32(0)

	//
	OpHandshakeReply = int32(1)

	// client 發送心跳
	OpHeartbeat = int32(2)

	// server 回覆心跳結果
	OpHeartbeatReply = int32(3)

	//
	OpSendMsg = int32(4)

	//
	OpSendMsgReply = int32(5)

	//
	OpDisconnectReply = int32(6)

	// client要求連線到某一個房間
	OpAuth = int32(7)

	// server回覆連線到某一個房間結果
	OpAuthReply = int32(8)

	// server訊息推送給client
	OpRaw = int32(9)

	// 處理tcp資料
	OpProtoReady = int32(10)

	// tcp close連線
	OpProtoFinish = int32(11)

	// 更換房間
	OpChangeRoom = int32(12)

	// 回覆更換房間結果
	OpChangeRoomReply = int32(13)

	// user新增operation
	OpSub = int32(14)

	// 回覆user新增operation結果
	OpSubReply = int32(15)

	// user移除operation
	OpUnsub = int32(16)

	// 回覆user移除operation結果
	OpUnsubReply = int32(17)
)

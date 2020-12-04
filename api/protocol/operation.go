package protocol

const (
	// OpHandshake handshake
	OpHandshake = int32(0)
	// OpHandshakeReply handshake reply
	OpHandshakeReply = int32(1)

	// OpHeartbeat heartbeat
	OpHeartbeat = int32(2)
	// OpHeartbeatReply heartbeat reply
	OpHeartbeatReply = int32(3)

	// OpSendMsg send message.
	OpSendMsg = int32(4)
	// OpSendMsgReply  send message reply
	OpSendMsgReply = int32(5)

	// OpDisconnectReply disconnect reply
	OpDisconnectReply = int32(6)

	// OpAuth auth connnect
	OpAuth = int32(7)
	// OpAuthReply auth connect reply
	OpAuthReply = int32(8)

	// OpRaw raw message
	OpRaw = int32(9)

	// OpProtoReady proto ready
	OpProtoReady = int32(10)
	// OpProtoFinish proto finish
	OpProtoFinish = int32(11)

	// OpChangeRoom change room
	OpChangeRoom = int32(12)
	// OpChangeRoomReply change room reply
	OpChangeRoomReply = int32(13)

	// OpSub subscribe operation
	OpSub = int32(14)
	// OpSubReply subscribe operation
	OpSubReply = int32(15)

	// OpUnsub unsubscribe operation
	OpUnsub = int32(16)
	// OpUnsubReply unsubscribe operation reply
	OpUnsubReply = int32(17)
)

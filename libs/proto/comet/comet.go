package comet

type PushMsgArg struct {
	Key       string
	Ver       int32
	Operation int32
	Msg       []byte
}

type PushMsgsArg struct {
	Key        string
	Vers       []int32
	Operations []int32
	Msgs       [][]byte
}

type PushMsgsReply struct {
	Index int32
}

type MPushMsgArg struct {
	Keys      []string
	Ver       int32
	Operation int32
	Msg       []byte
}

type MPushMsgReply struct {
	Index int32
}

type MPushMsgsArg struct {
	Keys       []string
	Vers       []int32
	Operations []int32
	Msgs       [][]byte
}

type MPushMsgsReply struct {
	Index int32
}

type BoardcastArg struct {
	Ver       int32
	Operation int32
	Msg       []byte
}

type BoardcastRoomArg struct {
	Ver       int32
	Operation int32
	Msg       []byte
	RoomId    int32
	Ensure    bool
}

type RoomsReply struct {
	Rooms map[int32]bool
}

package proto

type ConnArg struct {
	Token  string
	Server int32
}

type ConnReply struct {
	Key    string
	RoomId int32
}

type DisconnArg struct {
	Key    string
	RoomId int32
}

type DisconnReply struct {
	Has bool
}

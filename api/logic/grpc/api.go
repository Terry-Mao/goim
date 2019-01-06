package grpc

import (
	"fmt"
)

func (m *ConnectReq) String() string {
	return fmt.Sprintf("server:%s token:%s", m.Server, m.Token)
}

func (m *OnlineReq) String() string {
	return fmt.Sprintf("server:%s", m.Server)
}

func (m *OnlineReply) String() string {
	return fmt.Sprintf("rooms:%d", len(m.AllRoomCount))
}

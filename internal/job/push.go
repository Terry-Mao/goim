package job

import (
	"context"
	"fmt"

	pb "github.com/Terry-Mao/goim/api/logic/grpc"
)

func (j *Job) push(ctx context.Context, pushMsg *pb.PushMsg) (err error) {
	switch pushMsg.Type {
	case pb.PushMsg_PUSH:
	case pb.PushMsg_ROOM:
	case pb.PushMsg_BROADCAST:
	default:
		err = fmt.Errorf("no match type: %s", pushMsg.Type)
	}
	return
}

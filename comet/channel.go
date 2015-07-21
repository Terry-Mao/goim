package main

const (
	signalNum  = 1
	ProtoFinsh = 0
	ProtoReady = 1
)

// Channel used by message pusher send msg to write goroutine.
type Channel struct {
	Signal   chan int
	CliProto Ring
	SvrProto Ring
	//next     *Channel
}

func NewChannel(cliProto, svrProto int) *Channel {
	c := new(Channel)
	c.Signal = make(chan int, signalNum)
	InitRing(&c.CliProto, cliProto)
	InitRing(&c.SvrProto, svrProto)
	return c
}

// not goroutine safe, must push one by one.
func (c *Channel) PushMsg(ver int16, operation int32, body []byte) (err error) {
	var proto *Proto
	// fetch a proto from channel free list
	proto, err = c.SvrProto.Set()
	if err != nil {
		return
	}
	proto.Ver = ver
	proto.Operation = operation
	proto.Body = body
	c.SvrProto.SetAdv()
	select {
	case c.Signal <- ProtoReady:
	default:
	}
	return
}

// not goroutine safe, must push one by one.
func (c *Channel) PushMsgs(ver []int32, operations []int32, bodies [][]byte) (idx int32, err error) {
	var (
		proto *Proto
		n     int32
	)
	for n = 0; n < int32(len(ver)); n++ {
		// fetch a proto from channel free list
		if proto, err = c.SvrProto.Set(); err != nil {
			goto finish
		}
		proto.Ver = int16(ver[n])
		proto.Operation = operations[n]
		proto.Body = bodies[n]
		c.SvrProto.SetAdv()
		idx = n
	}
finish:
	select {
	case c.Signal <- ProtoReady:
	default:
	}
	return
}

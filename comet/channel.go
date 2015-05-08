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

func (c *Channel) Push(ver int16, operation int32, body []byte) (err error) {
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

func (c *Channel) Pushs(ver int16, operations []int32, bodies [][]byte) (n int, err error) {
	var (
		proto *Proto
	)
	if len(operations) != len(bodies) {
		err = ErrPushArgs
		return
	}
	for n = 0; n < len(operations); n++ {
		// fetch a proto from channel free list
		proto, err = c.SvrProto.Set()
		if err != nil {
			return
		}
		proto.Ver = ver
		proto.Operation = operations[n]
		proto.Body = bodies[n]
		c.SvrProto.SetAdv()
	}
	select {
	case c.Signal <- ProtoReady:
	default:
	}
	return
}

/*
func (c *Channel) Reset() {
	select {
	case <-c.Signal:
	default:
	}
	c.CliProto.Reset()
	c.SvrProto.Reset()
}
*/

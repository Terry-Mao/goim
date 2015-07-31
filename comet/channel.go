package main

const (
	signalNum   = 1
	protoFinish = 0
	protoReady  = 1
)

// Channel used by message pusher send msg to write goroutine.
type Channel struct {
	signal   chan int
	CliProto Ring
	SvrProto Ring
	//next     *Channel
}

func NewChannel(cliProto, svrProto int) *Channel {
	c := new(Channel)
	c.signal = make(chan int, signalNum)
	InitRing(&c.CliProto, cliProto)
	InitRing(&c.SvrProto, svrProto)
	return c
}

func (c *Channel) Ready() bool {
	return (<-c.signal) == protoReady
}

func (c *Channel) Signal() {
	// just ignore duplication signal
	select {
	case c.signal <- protoReady:
	default:
	}
}

func (c *Channel) Finish() {
	// don't use close chan, Signal can be reused
	// if chan full, writer goroutine next fetch from chan will exit
	// if chan empty, send a 0(close) let the writer exit
	select {
	case c.signal <- protoFinish:
	default:
	}
}

// not goroutine safe, must push one by one.
func (c *Channel) PushMsg(ver int16, operation int32, body []byte) (err error) {
	var proto *Proto
	// fetch a proto from channel free list
	if proto, err = c.SvrProto.Set(); err != nil {
		return
	}
	proto.Ver = ver
	proto.Operation = operation
	proto.Body = body
	c.SvrProto.SetAdv()
	c.Signal()
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
	c.Signal()
	return
}

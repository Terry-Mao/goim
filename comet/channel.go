package main

const (
	signalNum = 1
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

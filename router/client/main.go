package main

import (
	log "code.google.com/p/log4go"
	"flag"
	"fmt"
	"math"
	"net/rpc"
	"time"
)

const (
	TEST_NUM = 1

	SUB_DATA_NUM   = 2048000
	TOPIC_DATA_NUM = 512

	SUB_KEY   = "sub_id_Terry-Mao_%d"
	TOPIC_KEY = "topic_id_Felix-Hao_%d"

	RUN_TIME = 60
)

var (
	addr string

	stopCh = make(chan struct{}, 3)
)

type RPCTopicSubArg struct {
	Topic  string
	Subkey string
}

type RPCSubArg struct {
	Subkey string
	State  int8
	Server int16
}

type RPCSubMsg struct {
	Ret    int
	State  int8
	Server int16
}

func init() {
	flag.StringVar(&addr, "a", "127.0.0.1:9090", "rpc address")
}

func main() {
	flag.Parse()
	initData()
	testSub()
	testBatchSub()
	testTopic()
	time.Sleep(10 * time.Second)
	<-stopCh
	<-stopCh
	<-stopCh
}

func stop(bus string, counts []int64) {
	time.AfterFunc(RUN_TIME*time.Second, func() {
		c := int64(0)
		for _, v := range counts {
			c += v
		}
		log.Info("test %s stop, count %d, per %d", bus, c, c/RUN_TIME)
		stopCh <- struct{}{}
	})
}

func testSub() {
	c, err := rpc.Dial("tcp", addr)
	if err != nil {
		log.Error("rpc.Dial error(%v)", err)
		return
	}
	counts := make([]int64, TEST_NUM)
	for i := 0; i < TEST_NUM; i++ {
		go func(count *int64) {
			ret := &RPCSubMsg{}
			for i := 0; i < math.MaxInt64; i++ {
				key := fmt.Sprintf(SUB_KEY, i)
				c.Call("RouterRPC.Sub", &key, ret)
				*count++
			}
		}(&counts[i])
	}
	stop("sub", counts)
}

func testBatchSub() {
	c, err := rpc.Dial("tcp", addr)
	if err != nil {
		log.Error("rpc.Dial error(%v)", err)
		return
	}
	counts := make([]int64, TEST_NUM)
	for i := 0; i < TEST_NUM; i++ {
		go func(count *int64) {
			const (
				keys_num = 50
			)
			ret := &RPCSubMsg{}
			keys := make([]string, keys_num)
			for i := 0; i < math.MaxInt64; i++ {
				k := 0
				for j := i; j < i+keys_num; j++ {
					keys[k] = fmt.Sprintf(SUB_KEY, j)
					k++
				}
				c.Call("RouterRPC.BatchSub", &keys, ret)
				*count++
			}
		}(&counts[i])
	}
	stop("batch sub", counts)
}

func testTopic() {
	c, err := rpc.Dial("tcp", addr)
	if err != nil {
		log.Error("rpc.Dial error(%v)", err)
		return
	}
	counts := make([]int64, TEST_NUM)
	for i := 0; i < TEST_NUM; i++ {
		go func(count *int64) {
			ret := &RPCSubMsg{}
			for i := 0; i < math.MaxInt64; i++ {
				key := fmt.Sprintf(TOPIC_KEY, i)
				c.Call("RouterRPC.Topic", &key, ret)
				*count++
			}
		}(&counts[i])
	}
	stop("topic", counts)
}

func initData() {
	c, err := rpc.Dial("tcp", addr)
	if err != nil {
		log.Error("rpc.Dial error(%v)", err)
		return
	}
	reply := 0
	sb := &RPCSubArg{}
	sb.State = 1
	sb.Server = 1
	for i := 0; i < SUB_DATA_NUM; i++ {
		subkey := fmt.Sprintf(SUB_KEY, i)
		sb.Subkey = subkey
		c.Call("RouterRPC.SetSub", sb, &reply)
	}

	ts := &RPCTopicSubArg{}
	for i := 0; i < TOPIC_DATA_NUM; i++ {
		topic := fmt.Sprintf(TOPIC_KEY, i)
		ts.Topic = topic
		for j := 0; j < TOPIC_DATA_NUM*(i+1); j++ {
			ts.Subkey = fmt.Sprintf(SUB_KEY, j)
			c.Call("RouterRPC.SetTopic", ts, &reply)
		}
	}
	c.Close()
}

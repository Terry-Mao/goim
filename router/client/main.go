package main

import (
	log "code.google.com/p/log4go"
	"flag"
	"fmt"
	pbcodec "github.com/felixhao/goim/router/protobuf"
	"math"
	"net/rpc"
	"strings"
	"time"
)

const (
	TEST_NUM = 100

	SUB_DATA_NUM   = 2048000
	TOPIC_DATA_NUM = 128

	SUB_KEY   = "sub_id_Terry-Mao_%d"
	TOPIC_KEY = "topic_id_Felix-Hao_%d"

	RUN_TIME = 60
)

var (
	addrs string

	addrArr []string

	stopCh = make(chan struct{}, 4)
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
	flag.StringVar(&addrs, "a", "10.57.65.26:9090,10.57.65.26:9091,10.57.65.26:9092,10.57.65.26:9093", "rpc address")
}

func main() {
	flag.Parse()
	addrArr = strings.Split(addrs, ",")

	// initData()
	initPbData()
	// testSub()
	// testBatchSub()
	// testTopic()
	testPbSub()
	time.Sleep(10 * time.Second)
	<-stopCh
	// <-stopCh
	// <-stopCh
	// <-stopCh
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
	counts := make([]int64, len(addrArr)*TEST_NUM)
	ci := 0
	for _, v := range addrArr {
		for i := 0; i < TEST_NUM; i++ {
			go func(count *int64) {
				c, err := rpc.Dial("tcp", v)
				if err != nil {
					log.Error("rpc.Dial error(%v)", err)
					return
				}
				ret := &RPCSubMsg{}
				for i := 0; i < math.MaxInt64; i++ {
					key := fmt.Sprintf(SUB_KEY, i)
					c.Call("RouterRPC.Sub", &key, ret)
					*count++
				}
			}(&counts[ci])
			ci++
		}
	}
	stop("sub", counts)
}

func testBatchSub() {
	counts := make([]int64, len(addrArr)*TEST_NUM)
	ci := 0
	for _, v := range addrArr {
		for i := 0; i < TEST_NUM; i++ {
			go func(count *int64) {
				const (
					keys_num = 50
				)
				c, err := rpc.Dial("tcp", v)
				if err != nil {
					log.Error("rpc.Dial error(%v)", err)
					return
				}
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
			}(&counts[ci])
			ci++
		}
	}
	stop("batch sub", counts)
}

func testTopic() {
	counts := make([]int64, len(addrArr)*TEST_NUM)
	ci := 0
	for _, v := range addrArr {
		for i := 0; i < TEST_NUM; i++ {
			go func(count *int64) {
				c, err := rpc.Dial("tcp", v)
				if err != nil {
					log.Error("rpc.Dial error(%v)", err)
					return
				}
				ret := &RPCSubMsg{}
				for i := 0; i < math.MaxInt64; i++ {
					key := fmt.Sprintf(TOPIC_KEY, i)
					c.Call("RouterRPC.Topic", &key, ret)
					*count++
				}
			}(&counts[ci])
			ci++
		}
	}
	stop("topic", counts)
}

func testPbSub() {
	counts := make([]int64, len(addrArr)*TEST_NUM)
	ci := 0
	for _, v := range addrArr {
		for i := 0; i < TEST_NUM; i++ {
			go func(count *int64) {
				c, err := pbcodec.Dial("tcp", v)
				if err != nil {
					log.Error("rpc.Dial error(%v)", err)
					return
				}
				key := &PbRPCSubKey{}
				ret := &PbRPCSubRet{}
				for i := 0; i < math.MaxInt64; i++ {
					key.Key = fmt.Sprintf(TOPIC_KEY, i)
					c.Call("RouterRPC.PbSub", key, ret)
					*count++
				}
			}(&counts[ci])
			ci++
		}
	}
	stop("topic", counts)
}

func initData() {
	if len(addrArr) == 0 {
		panic("addrs is empty")
	}
	c, err := rpc.Dial("tcp", addrArr[0])
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
		for j := 0; j < i/2; j++ {
			ts.Subkey = fmt.Sprintf(SUB_KEY, j)
			c.Call("RouterRPC.SetTopic", ts, &reply)
		}
	}
	c.Close()
}

func initPbData() {
	if len(addrArr) == 0 {
		panic("addrs is empty")
	}
	c, err := pbcodec.Dial("tcp", addrArr[0])
	if err != nil {
		log.Error("rpc.Dial error(%v)", err)
		return
	}
	sb := &PbRPCSetSubArg{}
	sb.State = 1
	sb.Server = 1
	ret := &PbRPCSubRet{}
	for i := 0; i < SUB_DATA_NUM; i++ {
		subkey := fmt.Sprintf(SUB_KEY, i)
		sb.Subkey = subkey
		c.Call("RouterRPC.PbSetSub", sb, ret)
	}
	c.Close()
}

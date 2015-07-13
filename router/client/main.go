package main

import (
	log "code.google.com/p/log4go"
	"flag"
	"fmt"
	rpc "github.com/Terry-Mao/goim/protorpc"
	"github.com/Terry-Mao/goim/router/proto"
	"math"
	"strings"
	"sync"
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
	isID  bool
	addrs string

	addrArr []string

	stopCh = make(chan struct{}, 3)
)

func init() {
	flag.BoolVar(&isID, "i", false, "is or not init data")
	flag.StringVar(&addrs, "a", "127.0.0.1:9090", "rpc address")
}

func main() {
	flag.Parse()
	addrArr = strings.Split(addrs, ",")
	if isID {
		initData()
	}
	log.Info("start...")
	testSub()
	// testBatchSub()
	// testTopic()
	<-stopCh
	time.Sleep(5 * time.Second)
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
		log.Info("addr: %s", v)
		for i := 0; i < TEST_NUM; i++ {
			go func(count *int64) {
				c, err := rpc.Dial("tcp", v)
				if err != nil {
					log.Error("rpc.Dial error(%v)", err)
					return
				}
				key := &proto.ArgKey{}
				ret := &proto.Ret{}
				for i := 0; i < math.MaxInt64; i++ {
					key.Key = fmt.Sprintf(TOPIC_KEY, i)
					c.Call("RouterRPC.Sub", key, ret)
					*count++
					//log.Info("c: %d", *count)
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
				key := &proto.ArgBatchKey{Keys: make([]string, keys_num)}
				ret := &proto.RetBatchSub{}
				for i := 0; i < math.MaxInt64; i++ {
					k := 0
					for j := i; j < i+keys_num; j++ {
						key.Keys[k] = fmt.Sprintf(SUB_KEY, j)
						k++
					}
					c.Call("RouterRPC.BatchSub", key, ret)
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
				key := &proto.ArgTopic{}
				ret := &proto.RetBatchSub{}
				for i := 0; i < math.MaxInt64; i++ {
					key.Topic = fmt.Sprintf(TOPIC_KEY, i)
					c.Call("RouterRPC.Topic", key, ret)
					*count++
				}
			}(&counts[ci])
			ci++
		}
	}
	stop("topic", counts)
}

func initData() {
	al := len(addrArr)
	if al == 0 {
		panic("addrs is empty")
	}
	var wg sync.WaitGroup
	for i := 0; i < al; i++ {
		wg.Add(1)
		go func(n int) {
			c, err := rpc.Dial("tcp", addrArr[n])
			if err != nil {
				log.Error("rpc.Dial error(%v)", err)
				return
			}
			sb := &proto.ArgSub{}
			sb.State = 1
			sb.Server = 1
			ret := &proto.Ret{}
			e := SUB_DATA_NUM / al
			s := n * e
			for i := s; i < s+e; i++ {
				sb.Key = fmt.Sprintf(SUB_KEY, i)
				if err := c.Call("RouterRPC.SetSub", sb, ret); err != nil {
					panic(err)
				}
			}
			c.Close()
			wg.Done()
		}(i)
	}
	wg.Wait()
}

package main

import (
	"encoding/binary"
	"flag"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"

	log "code.google.com/p/log4go"
	"github.com/Shopify/sarama"
	cproto "github.com/Terry-Mao/goim/proto/comet"
	rproto "github.com/Terry-Mao/goim/proto/router"
	"github.com/Terry-Mao/protorpc"
	"github.com/wvanbergen/kafka/consumergroup"
)

//start eg: ./push_job >> push_job.log

const (
	KAFKA_GROUP_NAME                   = "kafka_topic_push_group"
	PUSH_MAX_BLOCK                     = 1000
	OFFSETS_PROCESSING_TIMEOUT_SECONDS = 10 * time.Second
	OFFSETS_COMMIT_INTERVAL            = 10 * time.Second

	OP_SEND_SMS_REPLY = int32(5)
)

func main() {
	flag.Parse()
	if err := InitConfig(); err != nil {
		panic(err)
	}

	log.LoadConfiguration(Conf.Log)
	if err := InitRouterRpc(Conf.RouterAddrs); err != nil {
		panic(err)
	}
	if err := InitCometRpc(Conf.Comets); err != nil {
		panic(err)
	}

	log.Info("start topic:%s consumer", Conf.KafkaTopic)
	runtime.GOMAXPROCS(runtime.NumCPU())

	log.Info("consumer group name:%s", KAFKA_GROUP_NAME)

	config := consumergroup.NewConfig()
	config.Offsets.Initial = sarama.OffsetNewest
	config.Offsets.ProcessingTimeout = OFFSETS_PROCESSING_TIMEOUT_SECONDS
	config.Offsets.CommitInterval = OFFSETS_COMMIT_INTERVAL
	config.Zookeeper.Chroot = Conf.ZKRoot

	kafkaTopics := []string{Conf.KafkaTopic}

	cg, err := consumergroup.JoinConsumerGroup(KAFKA_GROUP_NAME, kafkaTopics, Conf.ZKAddrs, config)
	if err != nil {
		panic(err)
		return
	}

	go func() {
		for err := range cg.Errors() {
			log.Error("consumer error(%v)", err)
		}
	}()

	go run(cg)

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT, syscall.SIGSTOP)
	for {
		s := <-c
		log.Info("get a signal %s\n", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGSTOP, syscall.SIGINT:
			if err := cg.Close(); err != nil {
				log.Error("Error closing the consumer error(%v)", err)
			}
			time.Sleep(3 * time.Second)
			log.Warn("consumer exit\n")
			return
		case syscall.SIGHUP:
			// TODO reload
		default:
			return
		}
	}
}

// run consume msg.
func run(cg *consumergroup.ConsumerGroup) {
	for msg := range cg.Messages() {
		log.Info("begin deal topic:%s, partitionId:%d, Offset:%d", msg.Topic, msg.Partition, msg.Offset)
		// key eg: cid, value eg: proto.GetReply
		if err := push(int64(binary.BigEndian.Uint64(msg.Key)), msg.Value); err != nil {
			log.Error("push(\"%s\") error(%v), try again", string(msg.Key), err)
		} else {
			log.Info("end delt success, topic:%s, Offset:%d, Key:%d", msg.Topic, msg.Offset, int64(binary.BigEndian.Uint64(msg.Key)))
		}
		cg.CommitUpto(msg)
	}
}

func push(userID int64, msg []byte) (err error) {
	var reply *rproto.GetReply
	reply, err = getSeqs(userID)
	if err != nil {
		log.Error("getSeqs(%d) error(%s)", userID, err)
		return
	}
	//分机器,推送
	m := divideComet(strconv.FormatInt(userID, 10), reply)
	for serverID, seqs := range m {
		c, ok := cometServiceMap[serverID]
		if !ok {
			log.Error("no serverID:%d, then ignore", serverID)
			continue
		}
		log.Debug("push to comet serverID:%d", serverID)

		i := 0
		wg := sync.WaitGroup{}
		loop := len(seqs) / PUSH_MAX_BLOCK
		wg.Add(loop + 1)
		for i = 0; i < loop; i++ {
			go pushToComet(serverID, c, seqs[i*PUSH_MAX_BLOCK:(i+1)*PUSH_MAX_BLOCK], msg, &wg)
		}
		go pushToComet(serverID, c, seqs[i*PUSH_MAX_BLOCK:], msg, &wg)
		wg.Wait()
	}
	return
}

func getSeqs(userID int64) (p *rproto.GetReply, err error) {
	arg := &rproto.GetArg{UserId: userID}
	p = &rproto.GetReply{}
	c := getRouterClient(userID)
	if err = c.Call(routerServiceGet, arg, p); err != nil {
		log.Error("c.Call(\"%s\", %v, reply) error(%v)", routerServiceGet, *arg, err)
		return
	}
	return
}

func divideComet(userID string, data *rproto.GetReply) map[int32][]string {
	var (
		subkey string
		m      = make(map[int32][]string, len(data.Seqs))
	)
	for i := 0; i < len(data.Seqs); i++ {
		subkey = userID + "_" + strconv.FormatInt(int64(data.Seqs[i]), 10)
		subkeys, ok := m[data.Servers[i]]
		if !ok {
			m[data.Servers[i]] = []string{subkey}
		} else {
			subkeys = append(subkeys, subkey)
			m[data.Servers[i]] = subkeys
		}
	}
	return m
}

func pushToComet(serverID int32, c *protorpc.Client, subkeys []string, body []byte, wg *sync.WaitGroup) {
	now := time.Now()
	args := &cproto.MPushMsgArg{Keys: subkeys, Operation: OP_SEND_SMS_REPLY, Msg: body}
	rep := &cproto.MPushMsgReply{}
	if err := c.Call(CometServiceMPushMsg, args, rep); err != nil {
		log.Error("c.Call(\"%s\", %v, reply) error(%v)", CometServiceMPushMsg, *args, err)
		goto done
	}
	log.Info("push msg to serverID:%d index:%d", serverID, rep.Index)
	log.Debug("push seconds %f", time.Now().Sub(now).Seconds())
done:
	wg.Done()
}

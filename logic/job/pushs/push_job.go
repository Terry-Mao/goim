package main

import (
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	log "code.google.com/p/log4go"
	"github.com/Shopify/sarama"
	cproto "github.com/Terry-Mao/goim/proto/comet"
	"github.com/Terry-Mao/protorpc"
	"github.com/wvanbergen/kafka/consumergroup"
)

//start eg: ./push_job > push_job.log

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

type KafkaPushs struct {
	Subkeys []string `json:"subkeys"`
	Msg     []byte   `json:"msg"`
}

// run consume msg.
func run(cg *consumergroup.ConsumerGroup) {
	for msg := range cg.Messages() {
		log.Info("begin deal topic:%s, partitionId:%d, Offset:%d", msg.Topic, msg.Partition, msg.Offset)
		// key eg: cid, value eg: proto.GetReply
		if err := push(int32(binary.BigEndian.Uint64(msg.Key)), msg.Value); err != nil {
			log.Error("push(\"%d\") error(%v), try again", int32(binary.BigEndian.Uint64(msg.Key)), err)
		} else {
			log.Info("end delt success, topic:%s, Offset:%d, Key:%d", msg.Topic, msg.Offset, int64(binary.BigEndian.Uint64(msg.Key)))
		}
		cg.CommitUpto(msg)
	}
}

func push(serverId int32, msg []byte) (err error) {
	tmp := KafkaPushs{}
	if err = json.Unmarshal(msg, &tmp); err != nil {
		log.Error("json.Unmarshal(%s) serverId:%d error(%s)", msg, serverId, err)
		return
	}

	c, ok := cometServiceMap[serverId]
	if !ok {
		err = fmt.Errorf("no serverId:%d, then ignore", serverId)
		log.Error(err)
		return
	}
	log.Debug("push to comet serverId:%d", serverId)

	i := 0
	wg := sync.WaitGroup{}
	loop := len(tmp.Subkeys) / PUSH_MAX_BLOCK
	wg.Add(loop + 1)
	for i = 0; i < loop; i++ {
		go pushToComet(serverId, c, tmp.Subkeys[i*PUSH_MAX_BLOCK:(i+1)*PUSH_MAX_BLOCK], tmp.Msg, &wg)
	}
	go pushToComet(serverId, c, tmp.Subkeys[i*PUSH_MAX_BLOCK:], tmp.Msg, &wg)
	wg.Wait()

	return
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

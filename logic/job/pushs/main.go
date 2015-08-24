package main

import (
	"encoding/json"
	"flag"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	log "code.google.com/p/log4go"
	"github.com/Shopify/sarama"
	cproto "github.com/Terry-Mao/goim/proto/comet"
	rproto "github.com/Terry-Mao/goim/proto/router"
	"github.com/Terry-Mao/protorpc"
	"github.com/thinkboy/goim/define"
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
	if err := InitRouter(); err != nil {
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
		// key eg  : message type
		// value eg: KafkaPushsMsg
		if err := push(string(msg.Key), msg.Value); err != nil {
			log.Error("push(\"%d\") error(%v), try again", msg.Key, err)
		} else {
			log.Info("end delt success, topic:%s, Offset:%d, Key:%s msg:%s", msg.Topic, msg.Offset, msg.Key, msg.Value)
		}
		cg.CommitUpto(msg)
	}
}

func push(op string, msg []byte) (err error) {
	if op == define.KAFKA_MESSAGE_MULTI {
		tmp := define.KafkaPushsMsg{}
		if err = json.Unmarshal(msg, &tmp); err != nil {
			log.Error("json.Unmarshal(%s) serverId:%d error(%s)", msg, err)
			return
		}
		go multiPush(tmp.UserIds, tmp.Msg)
	} else if op == define.KAFKA_MESSAGE_BROADCAST {
		broadcast(msg)
	}

	return
}

// multi-userids push
func multiPush(userIds []int64, msg []byte) {
	m := divideToRouter(userIds)    //m: map[router.serverId][]userId
	divide, err := divideToComet(m) //divide: map[comet.serverId][]subkey
	if err != nil {
		log.Error("divideToComet() error(%v)", err)
		return
	}
	for cometId, subkeys := range divide {
		c, err := getCometByServerId(cometId)
		if err != nil {
			log.Error("getCometByServerId(\"%d\") error(%v)", cometId, err)
			return
		}
		log.Debug("push to comet serverId:%d", cometId)
		i := 0
		loop := len(subkeys) / PUSH_MAX_BLOCK
		for i = 0; i < loop; i++ {
			go pushsMsgToComet(cometId, c, subkeys[i*PUSH_MAX_BLOCK:(i+1)*PUSH_MAX_BLOCK], msg)
		}
		go pushsMsgToComet(cometId, c, subkeys[i*PUSH_MAX_BLOCK:], msg)
	}
}

// mssage broadcast
func broadcast(msg []byte) {
	for serverId, c := range cometServiceMap {
		if *c == nil {
			log.Error("broadcast error(%v)", define.ErrComet)
			return
		}
		go broadcastToComet(serverId, *c, msg)
	}
}

// get subkeys from all routers and divide by comet-server-id
func divideToComet(m map[string][]int64) (divide map[int32][]string, err error) {
	divide = make(map[int32][]string) //map[comet.serverId][]subkey
	for routerId, us := range m {
		// TODO muti-routine get
		var reply *rproto.MGetReply
		reply, err = getSubkeys(routerId, us)
		if err != nil {
			log.Error("getSubkeys(\"%s\") error(%s)", routerId, err)
			return
		}
		log.Debug("getSubkeys:%v routerId:%s", reply.UserIds, routerId)
		for j := 0; j < len(reply.UserIds); j++ {
			s := reply.Sessions[j]
			log.Debug("sessions seqs:%v serverids:%v", s.Seqs, s.Servers)
			for i := 0; i < len(s.Seqs); i++ {
				subkey := define.Encode(reply.UserIds[j], s.Seqs[i])
				subkeys, ok := divide[s.Servers[i]]
				if !ok {
					subkeys = make([]string, 0, 1000) // TODO:consider
				}
				divide[s.Servers[i]] = append(subkeys, subkey)
			}
		}
	}
	return
}

func pushsMsgToComet(serverId int32, c *protorpc.Client, subkeys []string, body []byte) {
	now := time.Now()
	args := &cproto.MPushMsgArg{Keys: subkeys, Operation: OP_SEND_SMS_REPLY, Msg: body}
	rep := &cproto.MPushMsgReply{}
	if err := c.Call(CometServiceMPushMsg, args, rep); err != nil {
		log.Error("c.Call(\"%s\", %v, reply) error(%v)", CometServiceMPushMsg, *args, err)
		return
	}
	log.Info("push msg to serverId:%d index:%d", serverId, rep.Index)
	log.Debug("push seconds %f", time.Now().Sub(now).Seconds())
}

func broadcastToComet(serverId int32, c *protorpc.Client, msg []byte) {
	now := time.Now()
	args := &cproto.BoardcastArg{Ver: 0, Operation: OP_SEND_SMS_REPLY, Msg: msg}
	if err := c.Call(CometServiceBroadcast, args, nil); err != nil {
		log.Error("c.Call(\"%s\", %v, reply) error(%v)", CometServiceBroadcast, *args, err)
		return
	}
	log.Info("broadcast msg to serverId:%d msg:%s", serverId, msg)
	log.Debug("push seconds %f", time.Now().Sub(now).Seconds())
}

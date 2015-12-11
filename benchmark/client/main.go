package main

// Start Commond eg: ./client 1 5000 localhost:8080
// first parameter：beginning userId
// second parameter: amount of clients
// third parameter: comet server ip

import (
	"bufio"
	log "code.google.com/p/log4go"
	"encoding/binary"
	"flag"
	"fmt"
	mrand "math/rand"
	"net"
	"os"
	"runtime"
	"strconv"
	"sync/atomic"
	"time"
)

const (
	OP_HANDSHARE        = int32(0)
	OP_HANDSHARE_REPLY  = int32(1)
	OP_HEARTBEAT        = int32(2)
	OP_HEARTBEAT_REPLY  = int32(3)
	OP_SEND_SMS         = int32(4)
	OP_SEND_SMS_REPLY   = int32(5)
	OP_DISCONNECT_REPLY = int32(6)
	OP_AUTH             = int32(7)
	OP_AUTH_REPLY       = int32(8)
	OP_TEST             = int32(254)
	OP_TEST_REPLY       = int32(255)
)

const (
	rawHeaderLen = uint16(16)
	heart        = 240 * time.Second //s
)

type Proto struct {
	PackLen   int32  // package length
	HeaderLen int16  // header length
	Ver       int16  // protocol version
	Operation int32  // operation for request
	SeqId     int32  // sequence number chosen by client
	Body      []byte // body
}

var (
	countDown int64
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	log.Global = log.NewDefaultLogger(log.DEBUG)
	flag.Parse()
	defer log.Close()
	begin, err := strconv.Atoi(os.Args[1])
	if err != nil {
		panic(err)
	}

	num, err := strconv.Atoi(os.Args[2])
	if err != nil {
		panic(err)
	}

	//time.AfterFunc(time.Duration(t)*time.Second, stop)
	go result()

	for i := begin; i < begin+num; i++ {
		key := fmt.Sprintf("%d", i)
		go startClient(key)
	}

	var exit chan bool
	<-exit
}

func result() {
	var lastTimes int64
	for {
		lastTimes = countDown
		time.Sleep(time.Second)
		fmt.Println(fmt.Sprintf("down:%d down/s:%d", countDown, countDown-lastTimes))
	}
}

func startClient(key string) {
	time.Sleep(time.Duration(mrand.Intn(30)) * time.Second)

	conn, err := net.Dial("tcp", os.Args[3])
	if err != nil {
		log.Error("net.Dial(\"%s\") error(%v)", os.Args[2], err)
		return
	}
	seqId := int32(0)
	wr := bufio.NewWriter(conn)
	rd := bufio.NewReader(conn)
	proto := new(Proto)
	proto.Ver = 1
	// auth
	// test handshake timeout
	// time.Sleep(time.Second * 31)
	proto.Operation = OP_AUTH
	proto.SeqId = seqId
	proto.Body = []byte(key)
	if err = tcpWriteProto(wr, proto); err != nil {
		log.Error("tcpWriteProto() error(%v)", err)
		return
	}
	if err = tcpReadProto(rd, proto); err != nil {
		log.Error("tcpReadProto() error(%v)", err)
		return
	}
	log.Debug("auth ok, proto: %v", proto)
	seqId++
	// writer
	go func() {
		proto1 := new(Proto)
		for {
			// heartbeat
			proto1.Operation = OP_HEARTBEAT
			proto1.SeqId = seqId
			proto1.Body = nil
			if err = tcpWriteProto(wr, proto1); err != nil {
				log.Error("tcpWriteProto() error(%v)", err)
				return
			}
			// test heartbeat
			time.Sleep(heart)
			seqId++
			// op_test
			/*proto1.Operation = OP_TEST
			proto1.SeqId = seqId
			if err = tcpWriteProto(wr, proto1); err != nil {
				log.Error("tcpWriteProto() error(%v)", err)
				return
			}
			seqId++
			time.Sleep(10000 * time.Millisecond)*/
		}
	}()
	// reader
	for {
		if err = tcpReadProto(rd, proto); err != nil {
			log.Error("tcpReadProto() error(%v)", err)
			return
		}
		if proto.Operation == OP_HEARTBEAT_REPLY {
			log.Debug("receive heartbeat")
			if err = conn.SetReadDeadline(time.Now().Add(heart + 60*time.Second)); err != nil {
				log.Error("conn.SetReadDeadline() error(%v)", err)
				return
			}
		} else if proto.Operation == OP_TEST_REPLY {
			log.Debug("body: %s", string(proto.Body))
		} else if proto.Operation == OP_SEND_SMS_REPLY {
			log.Info("msg: %s", string(proto.Body))
			atomic.AddInt64(&countDown, 1)
		}
	}
}

func tcpWriteProto(wr *bufio.Writer, proto *Proto) (err error) {
	// write
	if err = binary.Write(wr, binary.BigEndian, uint32(rawHeaderLen)+uint32(len(proto.Body))); err != nil {
		return
	}
	if err = binary.Write(wr, binary.BigEndian, rawHeaderLen); err != nil {
		return
	}
	if err = binary.Write(wr, binary.BigEndian, proto.Ver); err != nil {
		return
	}
	if err = binary.Write(wr, binary.BigEndian, proto.Operation); err != nil {
		return
	}
	if err = binary.Write(wr, binary.BigEndian, proto.SeqId); err != nil {
		return
	}
	if proto.Body != nil {
		log.Debug("cipher body: %v", proto.Body)
		if err = binary.Write(wr, binary.BigEndian, proto.Body); err != nil {
			return
		}
	}
	err = wr.Flush()
	return
}

func tcpReadProto(rd *bufio.Reader, proto *Proto) (err error) {
	var (
		packLen   int32
		headerLen int16
	)
	// read
	if err = binary.Read(rd, binary.BigEndian, &packLen); err != nil {
		return
	}
	log.Debug("packLen: %d", packLen)
	if err = binary.Read(rd, binary.BigEndian, &headerLen); err != nil {
		return
	}
	log.Debug("headerLen: %d", headerLen)
	if err = binary.Read(rd, binary.BigEndian, &proto.Ver); err != nil {
		return
	}
	log.Debug("ver: %d", proto.Ver)
	if err = binary.Read(rd, binary.BigEndian, &proto.Operation); err != nil {
		return
	}
	log.Debug("operation: %d", proto.Operation)
	if err = binary.Read(rd, binary.BigEndian, &proto.SeqId); err != nil {
		return
	}
	log.Debug("seqId: %d", proto.SeqId)
	var (
		n       = int(0)
		t       = int(0)
		bodyLen = int(packLen - int32(headerLen))
	)
	log.Debug("read body len: %d", bodyLen)
	if bodyLen > 0 {
		proto.Body = make([]byte, bodyLen)
		for {
			if t, err = rd.Read(proto.Body[n:]); err != nil {
				return
			}
			if n += t; n == bodyLen {
				break
			} else if n < bodyLen {
			} else {
			}
		}
	} else {
		proto.Body = nil
	}
	return
}

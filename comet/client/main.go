package main

import (
	"bufio"
	log "code.google.com/p/log4go"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/binary"
	"flag"
	"github.com/Terry-Mao/goim/libs/crypto/padding"
	"github.com/Terry-Mao/goim/libs/perf"
	"net"
	"runtime"
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
)

type Proto struct {
	PackLen   int32  // package length
	HeaderLen int16  // header length
	Ver       int16  // protocol version
	Operation int32  // operation for request
	SeqId     int32  // sequence number chosen by client
	Body      []byte // body
}

func main() {
	flag.Parse()
	if err := InitConfig(); err != nil {
		panic(err)
	}
	runtime.GOMAXPROCS(Conf.MaxProc)
	log.LoadConfiguration(Conf.Log)
	defer log.Close()
	perf.Init(Conf.PprofBind)
	if err := InitRSA(); err != nil {
		panic(err)
	}
	conn, err := net.Dial("tcp", Conf.TCPAddr)
	if err != nil {
		log.Error("net.Dial(\"%s\") error(%v)", Conf.TCPAddr, err)
		return
	}
	seqId := int32(0)
	wr := bufio.NewWriter(conn)
	rd := bufio.NewReader(conn)
	proto := new(Proto)
	proto.Ver = 1
	key := make([]byte, 32)
	// handshake
	if _, err = rand.Read(key); err != nil {
		panic(err)
	}
	proto.Operation = OP_HANDSHARE
	proto.SeqId = seqId
	proto.Body = key
	// aes
	var (
		block cipher.Block
	)
	if block, err = aes.NewCipher(key[:16]); err != nil {
		log.Error("aes.NewCipher() error(%v)", err)
		return
	}
	log.Debug("aes key: %x, iv: %x", key[:16], key[16:])
	ebm := cipher.NewCBCEncrypter(block, key[16:])
	dbm := cipher.NewCBCDecrypter(block, key[16:])
	// rsa
	if proto.Body, err = rsa.EncryptOAEP(sha256.New(), rand.Reader, RSAPub, proto.Body, nil); err != nil {
		return
	}
	if err = WriteProto(wr, proto); err != nil {
		log.Error("WriteProto() error(%v)", err)
		return
	}
	if err = ReadProto(rd, proto); err != nil {
		log.Error("ReadProto() error(%v)", err)
		return
	}
	log.Debug("handshake ok, proto: %v", proto)
	dbm.CryptBlocks(proto.Body, proto.Body)
	if proto.Body, err = padding.PKCS7.Unpadding(proto.Body, dbm.BlockSize()); err != nil {
		log.Error("Unpadding() error(%v)", err)
		return
	}
	log.Debug("handshake sessionid : %s", string(proto.Body))
	seqId++
	// auth
	// test handshake timeout
	// time.Sleep(time.Second * 31)
	proto.Operation = OP_AUTH
	proto.SeqId = seqId
	proto.Body = padding.PKCS7.Padding([]byte(Conf.SubKey), ebm.BlockSize())
	// user aes encrypt sub key
	ebm.CryptBlocks(proto.Body, proto.Body)
	if err = WriteProto(wr, proto); err != nil {
		log.Error("WriteProto() error(%v)", err)
		return
	}
	if err = ReadProto(rd, proto); err != nil {
		log.Error("ReadProto() error(%v)", err)
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
			if err = WriteProto(wr, proto1); err != nil {
				log.Error("WriteProto() error(%v)", err)
				return
			}
			// test heartbeat
			//time.Sleep(time.Second * 31)
			seqId++
			// op_test
			proto1.Operation = OP_TEST
			proto1.SeqId = seqId
			// use aes
			proto1.Body = padding.PKCS7.Padding([]byte("hello test"), block.BlockSize())
			ebm.CryptBlocks(proto1.Body, proto1.Body)
			if err = WriteProto(wr, proto1); err != nil {
				log.Error("WriteProto() error(%v)", err)
				return
			}
			seqId++
			time.Sleep(10000 * time.Millisecond)
		}
	}()
	// reader
	for {
		if err = ReadProto(rd, proto); err != nil {
			log.Error("ReadProto() error(%v)", err)
			return
		}
		if proto.Body != nil {
			dbm.CryptBlocks(proto.Body, proto.Body)
		}
		if proto.Operation == OP_HEARTBEAT_REPLY {
			if err = conn.SetReadDeadline(time.Now().Add(25 * time.Second)); err != nil {
				log.Error("conn.SetReadDeadline() error(%v)", err)
				return
			}
			log.Debug("receive heartbeat")
		} else if proto.Operation == OP_TEST_REPLY {
			log.Debug("receive test")
			bodyStr, err := padding.PKCS7.Unpadding(proto.Body, block.BlockSize())
			if err != nil {
				log.Error("pkcs7.Unpadding() error(%v)", err)
				return
			}
			log.Debug("body: %s", bodyStr)
		} else if proto.Operation == OP_SEND_SMS_REPLY {
			log.Debug("receive message")
			bodyStr, err := padding.PKCS7.Unpadding(proto.Body, block.BlockSize())
			if err != nil {
				log.Error("pkcs7.Unpadding() error(%v)", err)
				return
			}
			log.Debug("body: %s", bodyStr)
		}
	}
}

func WriteProto(wr *bufio.Writer, proto *Proto) (err error) {
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

func ReadProto(rd *bufio.Reader, proto *Proto) (err error) {
	// read
	if err = binary.Read(rd, binary.BigEndian, &proto.PackLen); err != nil {
		return
	}
	log.Debug("packLen: %d", proto.PackLen)
	if err = binary.Read(rd, binary.BigEndian, &proto.HeaderLen); err != nil {
		return
	}
	log.Debug("headerLen: %d", proto.HeaderLen)
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
	if err = ReadBody(rd, proto); err != nil {
	}
	return
}

func ReadBody(rd *bufio.Reader, proto *Proto) (err error) {
	var (
		n       = int(0)
		t       = int(0)
		bodyLen = int(proto.PackLen - int32(proto.HeaderLen))
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

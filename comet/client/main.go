package main

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"encoding/binary"
	"fmt"
	"github.com/Terry-Mao/goim/libs/crypto/aes"
	"github.com/Terry-Mao/goim/libs/crypto/padding"
	myrsa "github.com/Terry-Mao/goim/libs/crypto/rsa"
	"net"
	"time"
)

const (
	OP_HANDSHARE        = uint32(0)
	OP_HANDSHARE_REPLY  = uint32(1)
	OP_HEARTBEAT        = uint32(2)
	OP_HEARTBEAT_REPLY  = uint32(3)
	OP_SEND_SMS         = uint32(4)
	OP_SEND_SMS_REPLY   = uint32(5)
	OP_DISCONNECT_REPLY = uint32(6)
	OP_TEST             = uint32(254)
	OP_TEST_REPLY       = uint32(255)
)

const (
	packBytes    = 4
	headerBytes  = 2
	rawPackLen   = uint32(16)
	rawHeaderLen = uint16(12)

	rsaPubKey = `
-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQC0uoYIqecHK2c9CgyEKWDK5XGr
YLT29CgENUm9eBPi4YyCGCXqaesdRs1TS7X7JKpAh114BGkkNPuEEFHbzIgIHSoN
GIB9r/ustGGggdeqqFiEhq6vxWM85RPWBGxv3WNAnwVqZ+NJ5+1Q0Rwpaazr6wr6
LddByFzf/U88GQfzhQIDAQAB
-----END PUBLIC KEY-----
`
)

var (
	RSAPub *rsa.PublicKey
)

func init() {
	var err error
	if RSAPub, err = myrsa.PublicKey([]byte(rsaPubKey)); err != nil {
		panic(err)
	}
}

type Proto struct {
	Ver       uint16 // protocol version
	Operation uint32 // operation for request
	SeqId     uint32 // sequence number chosen by client
	Body      []byte // body
}

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		panic(err)
	}
	var body []byte
	seqId := uint32(0)
	wr := bufio.NewWriter(conn)
	rd := bufio.NewReader(conn)
	proto := new(Proto)
	proto.Ver = 1
	aesKey := make([]byte, 16)
	// handshake
	if _, err = rand.Read(aesKey); err != nil {
		panic(err)
	}
	fmt.Printf("aes key: %v\n", aesKey)
	proto.Operation = OP_HANDSHARE
	proto.SeqId = seqId
	proto.Body = aesKey
	// use rsa
	if proto.Body, err = myrsa.Encrypt(proto.Body, RSAPub); err != nil {
		return
	}
	if err = WriteProto(wr, proto); err != nil {
		panic(err)
	}
	if _, err = ReadProto(rd, proto); err != nil {
		panic(err)
	}
	fmt.Printf("handshake ok, proto: %v\n", proto)
	seqId++
	// writer
	go func() {
		proto1 := new(Proto)
		for {
			proto1.Operation = OP_HEARTBEAT
			proto1.SeqId = seqId
			proto1.Body = nil
			if err = WriteProto(wr, proto1); err != nil {
				panic(err)
			}
			proto1.Operation = OP_TEST
			proto1.SeqId = seqId
			// use aes
			if body, err = aes.ECBEncrypt([]byte("hello test"), aesKey, padding.PKCS5); err != nil {
				panic(err)
			}
			proto1.Body = body
			if err = WriteProto(wr, proto1); err != nil {
				panic(err)
			}
			time.Sleep(500 * time.Millisecond)
		}
	}()
	// reader
	for {
		if proto.Body, err = ReadProto(rd, proto); err != nil {
			panic(err)
		}
		if proto.Body != nil {
			if body, err = aes.ECBDecrypt(body, aesKey, padding.PKCS5); err != nil {
				panic(err)
			}
			fmt.Printf("body: %s\n", string(body))
		}
		if proto.Operation == OP_HEARTBEAT_REPLY {
			if err = conn.SetReadDeadline(time.Now().Add(25 * time.Second)); err != nil {
				panic(err)
			}
			fmt.Printf("receive heartbeat\n")
		}
		seqId++
	}
}

func WriteProto(wr *bufio.Writer, proto *Proto) (err error) {
	// write
	if err = binary.Write(wr, binary.BigEndian, rawPackLen+uint32(len(proto.Body))); err != nil {
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
		fmt.Printf("cipher body: %v\n", proto.Body)
		if err = binary.Write(wr, binary.BigEndian, proto.Body); err != nil {
			return
		}
	}
	err = wr.Flush()
	return
}

func ReadProto(rd *bufio.Reader, proto *Proto) (body []byte, err error) {
	var (
		packLen   uint32
		headerLen uint16
	)
	// read
	if err = binary.Read(rd, binary.BigEndian, &packLen); err != nil {
		return
	}
	fmt.Printf("packLen: %d\n", packLen)
	if err = binary.Read(rd, binary.BigEndian, &headerLen); err != nil {
		return
	}
	fmt.Printf("headerLen: %d\n", headerLen)
	if err = binary.Read(rd, binary.BigEndian, &proto.Ver); err != nil {
		return
	}
	fmt.Printf("ver: %d\n", proto.Ver)
	if err = binary.Read(rd, binary.BigEndian, &proto.Operation); err != nil {
		return
	}
	fmt.Printf("operation: %d\n", proto.Operation)
	if err = binary.Read(rd, binary.BigEndian, &proto.SeqId); err != nil {
		return
	}
	fmt.Printf("seqId: %d\n", proto.SeqId)
	if body, err = ReadBody(packLen, headerLen, rd); err != nil {
	}
	return
}

func ReadBody(packLen uint32, headerLen uint16, rd *bufio.Reader) (body []byte, err error) {
	var (
		n       = int(0)
		t       = int(0)
		bodyLen = int(packLen - uint32(headerLen) - packBytes)
	)
	fmt.Printf("read body len: %d\n", bodyLen)
	if bodyLen > 0 {
		body = make([]byte, bodyLen)
		for {
			if t, err = rd.Read(body[n:]); err != nil {
				return
			}
			if n += t; n == bodyLen {
				break
			} else if n < bodyLen {
			} else {
			}
		}
	} else {
		body = nil
	}
	return
}

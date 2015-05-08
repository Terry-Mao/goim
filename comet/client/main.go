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

	rsaPubKey = `
-----BEGIN PUBLIC KEY-----
MDcwDQYJKoZIhvcNAQEBBQADJgAwIwIcAN4ev8NwvaxP22fQv/fUGc8/uNnfjHde
CgqUPQIDAQAB
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
	PackLen   int32  // package length
	HeaderLen int16  // header length
	Ver       int16  // protocol version
	Operation int32  // operation for request
	SeqId     int32  // sequence number chosen by client
	Body      []byte // body
}

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		panic(err)
	}
	var body []byte
	seqId := int32(0)
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
		panic(err)
	}
	if err = WriteProto(wr, proto); err != nil {
		panic(err)
	}
	if err = ReadProto(rd, proto); err != nil {
		panic(err)
	}
	fmt.Printf("handshake ok, proto: %v\n", proto)
	seqId++
	// auth
	proto.Operation = OP_AUTH
	proto.SeqId = seqId
	proto.Body = []byte("Terry-Mao")
	// user aes encrypt sub key
	if proto.Body, err = aes.ECBEncrypt(proto.Body, aesKey, padding.PKCS5); err != nil {
		panic(err)
	}
	if err = WriteProto(wr, proto); err != nil {
		panic(err)
	}
	if err = ReadProto(rd, proto); err != nil {
		panic(err)
	}
	fmt.Printf("auth ok, proto: %v\n", proto)
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
			seqId++
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
			seqId++
			time.Sleep(10000 * time.Millisecond)
		}
	}()
	// reader
	for {
		if err = ReadProto(rd, proto); err != nil {
			panic(err)
		}
		if proto.Body != nil {
			if proto.Body, err = aes.ECBDecrypt(body, aesKey, padding.PKCS5); err != nil {
				panic(err)
			}
		}
		if proto.Operation == OP_HEARTBEAT_REPLY {
			if err = conn.SetReadDeadline(time.Now().Add(25 * time.Second)); err != nil {
				panic(err)
			}
			fmt.Printf("receive heartbeat\n")
		} else if proto.Operation == OP_TEST_REPLY {
			fmt.Printf("receive test\n")
			fmt.Printf("body: %s\n", string(proto.Body))
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
		fmt.Printf("cipher body: %v\n", proto.Body)
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
	fmt.Printf("packLen: %d\n", proto.PackLen)
	if err = binary.Read(rd, binary.BigEndian, &proto.HeaderLen); err != nil {
		return
	}
	fmt.Printf("headerLen: %d\n", proto.HeaderLen)
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
	fmt.Printf("read body len: %d\n", bodyLen)
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

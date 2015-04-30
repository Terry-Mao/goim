package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

const (
	OP_HANDSHARE       = uint32(0)
	OP_HEARTBEAT       = uint32(1)
	OP_HEARTBEAT_REPLY = uint32(2)
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		panic(err)
	}
	wr := bufio.NewWriter(conn)
	rd := bufio.NewReader(conn)
	packLen := uint32(0)
	headerLen := uint16(0)
	ver := uint16(0)
	op := OP_HEARTBEAT
	seqId := uint32(0)
	for {
		// write
		if err := binary.Write(wr, binary.BigEndian, uint32(16)); err != nil {
			panic(err)
		}
		if err := binary.Write(wr, binary.BigEndian, uint16(12)); err != nil {
			panic(err)
		}
		if err := binary.Write(wr, binary.BigEndian, uint16(1)); err != nil {
			panic(err)
		}
		if err := binary.Write(wr, binary.BigEndian, OP_HEARTBEAT); err != nil {
			panic(err)
		}
		if err := binary.Write(wr, binary.BigEndian, seqId); err != nil {
			panic(err)
		}
		if err := wr.Flush(); err != nil {
			panic(err)
		}
		fmt.Println("send heartbeat")
		seqId++
		// read
		if err := binary.Read(rd, binary.BigEndian, &packLen); err != nil {
			panic(err)
		}
		fmt.Printf("packLen: %d\n", packLen)
		if err := binary.Read(rd, binary.BigEndian, &headerLen); err != nil {
			panic(err)
		}
		fmt.Printf("packLen: %d\n", headerLen)
		if err := binary.Read(rd, binary.BigEndian, &ver); err != nil {
			panic(err)
		}
		fmt.Printf("ver: %d\n", ver)
		if err := binary.Read(rd, binary.BigEndian, &op); err != nil {
			panic(err)
		}
		fmt.Printf("op: %d\n", op)
		if err := binary.Read(rd, binary.BigEndian, &seqId); err != nil {
			panic(err)
		}
		fmt.Printf("seqId: %d\n", seqId)
		time.Sleep(10 * time.Second)
	}
}

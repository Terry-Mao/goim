// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package protorpc

import (
	"errors"
	"fmt"
	"github.com/gogo/protobuf/proto"
	"net"
	"runtime"
	"strings"
	"testing"
)

type shutdownCodec struct {
	responded chan int
	closed    bool
}

func (c *shutdownCodec) WriteRequest(*Request, proto.Message) error { return nil }
func (c *shutdownCodec) ReadResponseBody(proto.Message) error       { return nil }
func (c *shutdownCodec) ReadResponseHeader(*Response) error {
	c.responded <- 1
	return errors.New("shutdownCodec ReadResponseHeader")
}
func (c *shutdownCodec) Close() error {
	c.closed = true
	return nil
}

func TestCloseCodec(t *testing.T) {
	codec := &shutdownCodec{responded: make(chan int)}
	client := NewClientWithCodec(codec)
	<-codec.responded
	client.Close()
	if !codec.closed {
		t.Error("client.Close did not close codec")
	}
}

// Test that errors in protobuf shut down the connection. Issue 7689.

type R struct {
	msg []byte // Not exported, so R does not work with protobuf.
}

type S struct{}

func (s *S) Recv(nul proto.Message, reply *R) error {
	*reply = R{[]byte("foo")}
	return nil
}

func TestProtoError(t *testing.T) {
	if runtime.GOOS == "plan9" {
		t.Skip("skipping test; see http://golang.org/issue/8908")
	}
	defer func() {
		err := recover()
		if err == nil {
			t.Fatal("no error")
		}
		if !strings.Contains("reading body EOF", err.(error).Error()) {
			t.Fatal("expected `reading body EOF', got", err)
		}
	}()
	Register(new(S))

	listen, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	go Accept(listen)

	client, err := Dial("tcp", listen.Addr().String())
	if err != nil {
		panic(err)
	}

	var reply Reply
	err = client.Call("S.Recv", nil, &reply)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v\n", reply)
	client.Close()

	listen.Close()
}

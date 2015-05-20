package main

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"testing"
)

const (
	TEST_NUM = 5
)

func TestRouter(t *testing.T) {
	testBucket(t)
}

func testBucket(t *testing.T) {
	InitConfig()
	InitBuckets()
	bs := make([]byte, 16)
	topics := make([]string, TEST_NUM)
	subs := make([]string, TEST_NUM)
	for i := 0; i < TEST_NUM; i++ {
		io.ReadFull(rand.Reader, bs)
		subkey := hex.EncodeToString(bs)
		subs[i] = subkey
		io.ReadFull(rand.Reader, bs)
		topic := hex.EncodeToString(bs)
		topics[i] = topic

		if i%2 == 0 {
			DefaultBuckets.SubBucket(subkey).SerStateAndServer(subkey, 1, 1)
		} else {
			DefaultBuckets.SubBucket(subkey).SerStateAndServer(subkey, 1, 1)
			DefaultBuckets.PutToTopic(topic, subkey)
		}
		if i%3 == 0 {
			DefaultBuckets.DelFromTopic(topic, subkey)
			t.Log("del topic", topic, "sub", subkey)
		}
	}

	for i := 0; i < TEST_NUM; i++ {
		subkey := subs[i]
		n := DefaultBuckets.SubBucket(subkey).Get(subkey)
		t.Log("sub", subkey, n)

		topic := topics[i]
		ss := DefaultBuckets.TopicBucket(topic).Get(topic)
		t.Log("topic", topic, ss)
	}
}

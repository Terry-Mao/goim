package main

import (
	"github.com/Terry-Mao/goim/libs/hash/cityhash"
	"sync"
)

var (
	DefaultBuckets *Buckets
)

func InitBuckets() {
	DefaultBuckets = NewBuckets()
}

type Buckets struct {
	subBuckets     []*SubBucket
	subBucketIdx   uint32
	topicBuckets   []*TopicBucket
	topicBucketIdx uint32
}

func NewBuckets() *Buckets {
	s := new(Buckets)
	s.subBuckets = make([]*SubBucket, Conf.SubBucketNum)
	s.subBucketIdx = uint32(Conf.SubBucketNum - 1)
	for i := 0; i < Conf.SubBucketNum; i++ {
		s.subBuckets[i] = NewSubBucket()
	}
	s.topicBuckets = make([]*TopicBucket, Conf.TopicBucketNum)
	s.topicBucketIdx = uint32(Conf.TopicBucketNum - 1)
	for i := 0; i < Conf.TopicBucketNum; i++ {
		s.topicBuckets[i] = NewTopicBucket()
	}
	return s
}

func (s *Buckets) SubBucket(key string) *SubBucket {
	idx := cityhash.CityHash32([]byte(key), uint32(len(key))) & s.subBucketIdx
	// log.Debug("\"%s\" hit sub bucket index: %d use cityhash", key, idx)
	return s.subBuckets[idx]
}

func (s *Buckets) TopicBucket(key string) *TopicBucket {
	idx := cityhash.CityHash32([]byte(key), uint32(len(key))) & s.topicBucketIdx
	// log.Debug("\"%s\" hit topic bucket index: %d use cityhash", key, idx)
	return s.topicBuckets[idx]
}

func (s *Buckets) PutToTopic(topic, subkey string) {
	tb := s.TopicBucket(topic)
	if tb != nil {
		tb.put(topic, subkey)
	}
	sb := s.SubBucket(subkey)
	if sb != nil {
		sb.putTopic(subkey, topic)
	}
}

func (s *Buckets) DelFromTopic(topic, subkey string) {
	tb := s.TopicBucket(topic)
	if tb != nil {
		tb.del(topic, subkey)
	}
	sb := s.SubBucket(subkey)
	if sb != nil {
		sb.delTopic(subkey, topic)
	}
}

func (s *Buckets) RemoveTopic(topic string) {
	tb := s.TopicBucket(topic)
	if tb != nil {
		m := tb.remove(topic)
		if m != nil {
			for k, _ := range m {
				sb := s.SubBucket(k)
				if sb != nil {
					sb.delTopic(k, topic)
				}
			}
		}
	}
}

type Node struct {
	state  int8
	server int16
	topics map[string]struct{}
}

func NewNode() *Node {
	return new(Node)
}

type SubBucket struct {
	bLock sync.Mutex
	subs  map[string]*Node
}

func NewSubBucket() *SubBucket {
	b := new(SubBucket)
	b.subs = make(map[string]*Node)
	return b
}

func (sb *SubBucket) putTopic(subkey, topic string) {
	sb.bLock.Lock()
	n, ok := sb.subs[subkey]
	if !ok {
		n = NewNode()
	}
	if n.topics == nil {
		n.topics = make(map[string]struct{})
	}
	n.topics[topic] = struct{}{}
	sb.bLock.Unlock()
}

func (sb *SubBucket) delTopic(subkey, topic string) {
	sb.bLock.Lock()
	n, ok := sb.subs[subkey]
	if ok && n.topics != nil {
		delete(n.topics, topic)
		sb.subs[subkey] = n
	}
	sb.bLock.Unlock()
}

func (sb *SubBucket) SetState(subkey string, state int8) {
	sb.bLock.Lock()
	n, ok := sb.subs[subkey]
	if !ok {
		n = NewNode()
		sb.subs[subkey] = n
	}
	n.state = state
	sb.bLock.Unlock()
}

func (sb *SubBucket) SetServer(subkey string, server int16) {
	sb.bLock.Lock()
	n, ok := sb.subs[subkey]
	if !ok {
		n = NewNode()
		sb.subs[subkey] = n
	}
	n.server = server
	sb.bLock.Unlock()
}

func (sb *SubBucket) SetStateAndServer(subkey string, state int8, server int16) {
	sb.bLock.Lock()
	n, ok := sb.subs[subkey]
	if !ok {
		n = NewNode()
		sb.subs[subkey] = n
	}
	n.state = state
	n.server = server
	sb.bLock.Unlock()
}

func (b *SubBucket) Get(subkey string) *Node {
	var n *Node
	b.bLock.Lock()
	n = b.subs[subkey]
	b.bLock.Unlock()
	return n
}

type TopicBucket struct {
	tLock  sync.Mutex
	topics map[string]map[string]struct{}
}

func NewTopicBucket() *TopicBucket {
	tb := new(TopicBucket)
	tb.topics = make(map[string]map[string]struct{})
	return tb
}

func (tb *TopicBucket) put(topic, subkey string) {
	tb.tLock.Lock()
	l, ok := tb.topics[topic]
	if !ok {
		l = make(map[string]struct{})
	}
	l[subkey] = struct{}{}
	tb.topics[topic] = l
	tb.tLock.Unlock()
}

func (tb *TopicBucket) del(topic, subkey string) {
	tb.tLock.Lock()
	l, ok := tb.topics[topic]
	if ok {
		delete(l, subkey)
		tb.topics[topic] = l
	}
	tb.tLock.Unlock()
}

func (tb *TopicBucket) remove(topic string) map[string]struct{} {
	var m map[string]struct{}
	tb.tLock.Lock()
	m = tb.topics[topic]
	delete(tb.topics, topic)
	tb.tLock.Unlock()
	return m
}

func (tb *TopicBucket) Get(topic string) map[string]struct{} {
	var m map[string]struct{}
	tb.tLock.Lock()
	m = tb.topics[topic]
	tb.tLock.Unlock()
	return m
}

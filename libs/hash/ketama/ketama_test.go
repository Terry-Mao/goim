package ketama

import (
	"strconv"
	"testing"
)

func Benchmark_Hash(b *testing.B) {
	ring := NewRing(255)
	ring.AddNode("node1", 1)
	ring.AddNode("node2", 1)
	ring.AddNode("node3", 1)
	ring.AddNode("node4", 1)
	ring.AddNode("node5", 1)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		ring.Hash(strconv.Itoa(i))
	}
}

package ketama

import (
	"crypto/sha1"
	"sort"
	"strconv"
)

const (
	// TODO you can modify this get more virtual node
	Base = 255
)

type node struct {
	node string
	hash uint
}

type tickArray []node

func (p tickArray) Len() int           { return len(p) }
func (p tickArray) Less(i, j int) bool { return p[i].hash < p[j].hash }
func (p tickArray) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p tickArray) Sort()              { sort.Sort(p) }

type HashRing struct {
	defaultSpots int
	ticks        tickArray
	length       int
}

func NewRing(n int) (h *HashRing) {
	h = new(HashRing)
	h.defaultSpots = n
	return
}

// Adds a new node to a hash ring
// n: name of the server
// s: multiplier for default number of ticks (useful when one cache node has more resources, like RAM, than another)
func (h *HashRing) AddNode(n string, s int) {
	tSpots := h.defaultSpots * s
	hash := sha1.New()
	for i := 1; i <= tSpots; i++ {
		hash.Write([]byte(n + ":" + strconv.Itoa(i)))
		hashBytes := hash.Sum(nil)

		n := &node{
			node: n,
			hash: uint(hashBytes[19]) | uint(hashBytes[18])<<8 | uint(hashBytes[17])<<16 | uint(hashBytes[16])<<24,
		}

		h.ticks = append(h.ticks, *n)
		hash.Reset()
	}
}

func (h *HashRing) Bake() {
	h.ticks.Sort()
	h.length = len(h.ticks)
}

func (h *HashRing) Hash(s string) string {
	hash := sha1.New()
	hash.Write([]byte(s))
	hashBytes := hash.Sum(nil)
	v := uint(hashBytes[19]) | uint(hashBytes[18])<<8 | uint(hashBytes[17])<<16 | uint(hashBytes[16])<<24
	i := sort.Search(h.length, func(i int) bool { return h.ticks[i].hash >= v })

	if i == h.length {
		i = 0
	}

	return h.ticks[i].node
}

package comet

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"time"

	logic "github.com/Terry-Mao/goim/api/logic/grpc"
	"github.com/Terry-Mao/goim/internal/comet/conf"
	log "github.com/golang/glog"
	"github.com/google/uuid"
	"github.com/zhenjl/cityhash"
	"google.golang.org/grpc"
)

var (
	_maxInt = 1<<31 - 1
)

const (
	_clientHeartbeat       = time.Second * 90
	_minSrvHeartbeatSecond = 600  // 10m
	_maxSrvHeartbeatSecond = 1200 // 20m
)

// Server .
type Server struct {
	c         *conf.Config
	round     *Round    // accept round store
	buckets   []*Bucket // subkey bucket
	bucketIdx uint32

	serverID  string
	rpcClient logic.LogicClient
}

func newLogicClient(c *conf.RPCClient) logic.LogicClient {
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithTimeout(time.Second),
		grpc.WithCompressor(grpc.NewGZIPCompressor()),
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.Dial))
	defer cancel()
	conn, err := grpc.DialContext(ctx, "discovery://default/goim.logic", opts...)
	if err != nil {
		panic(err)
	}
	return logic.NewLogicClient(conn)
}

// NewServer returns a new Server.
func NewServer(c *conf.Config) *Server {
	s := &Server{
		c:         c,
		round:     NewRound(c),
		rpcClient: newLogicClient(c.RPCClient),
	}
	// init bucket
	s.buckets = make([]*Bucket, c.Bucket.Size)
	s.bucketIdx = uint32(c.Bucket.Size)
	for i := 0; i < c.Bucket.Size; i++ {
		s.buckets[i] = NewBucket(c.Bucket)
	}
	var err error
	if s.serverID, err = os.Hostname(); err != nil {
		u, _ := uuid.NewRandom()
		s.serverID = u.String()
	}
	go s.onlineproc()
	return s
}

// Buckets return all buckets.
func (s *Server) Buckets() []*Bucket {
	return s.buckets
}

// Bucket get the bucket by subkey.
func (s *Server) Bucket(subKey string) *Bucket {
	idx := cityhash.CityHash32([]byte(subKey), uint32(len(subKey))) % s.bucketIdx
	if conf.Conf.Debug {
		log.Infof("%s hit channel bucket index: %d use cityhash", subKey, idx)
	}
	return s.buckets[idx]
}

// NextKey generate a server key.
func (s *Server) NextKey() string {
	u, err := uuid.NewRandom()
	if err == nil {
		return u.String()
	}
	return fmt.Sprintf("%s-%d", s.serverID, time.Now().UnixNano())
}

// RandServerHearbeat rand server heartbeat.
func (s *Server) RandServerHearbeat() time.Duration {
	return time.Duration(_minSrvHeartbeatSecond+rand.Intn(_maxSrvHeartbeatSecond-_minSrvHeartbeatSecond)) * time.Second
}

// Close close the server.
func (s *Server) Close() (err error) {
	return
}

func (s *Server) onlineproc() {
	for {
		var (
			allRoomsCount map[string]int32
			err           error
		)
		roomCount := make(map[string]int32)
		for _, bucket := range s.buckets {
			for roomID, count := range bucket.RoomsCount() {
				roomCount[roomID] += count
			}
		}
		if allRoomsCount, err = s.RenewOnline(s.serverID, roomCount); err != nil {
			time.Sleep(time.Duration(s.c.OnlineTick))
			continue
		}
		for _, bucket := range s.buckets {
			bucket.UpRoomsCount(allRoomsCount)
		}
		time.Sleep(time.Duration(s.c.OnlineTick))
	}
}

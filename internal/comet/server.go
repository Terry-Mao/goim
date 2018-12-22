package comet

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	logic "github.com/Terry-Mao/goim/api/logic/grpc"
	"github.com/Terry-Mao/goim/internal/comet/conf"
	log "github.com/golang/glog"
	"github.com/google/uuid"
	"github.com/zhenjl/cityhash"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

var (
	maxInt = 1<<31 - 1
	// grpc options
	grpcKeepAliveTime    = time.Duration(10) * time.Second
	grpcKeepAliveTimeout = time.Duration(3) * time.Second
	grpcBackoffMaxDelay  = time.Duration(3) * time.Second
	grpcMaxSendMsgSize   = 1 << 24
	grpcMaxCallMsgSize   = 1 << 24
)

const (
	clientHeartbeat       = time.Second * 90
	minSrvHeartbeatSecond = time.Minute * 10
	maxSrvHeartbeatSecond = time.Minute * 30
	// grpc options
	grpcInitialWindowSize     = 1 << 24
	grpcInitialConnWindowSize = 1 << 24
)

// Server is comet server.
type Server struct {
	c         *conf.Config
	round     *Round    // accept round store
	buckets   []*Bucket // subkey bucket
	bucketIdx uint32

	serverID  string
	rpcClient logic.LogicClient
}

func newLogicClient(c *conf.RPCClient) logic.LogicClient {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.Dial))
	defer cancel()
	conn, err := grpc.DialContext(ctx, "discovery://default/goim.logic",
		[]grpc.DialOption{
			grpc.WithInsecure(),
			grpc.WithInitialWindowSize(grpcInitialWindowSize),
			grpc.WithInitialConnWindowSize(grpcInitialConnWindowSize),
			grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(grpcMaxCallMsgSize)),
			grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(grpcMaxSendMsgSize)),
			grpc.WithBackoffMaxDelay(grpcBackoffMaxDelay),
			grpc.WithKeepaliveParams(keepalive.ClientParameters{
				Time:                grpcKeepAliveTime,
				Timeout:             grpcKeepAliveTimeout,
				PermitWithoutStream: true,
			}),
		}...)
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
	s.serverID = c.Env.Host
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
	return (minSrvHeartbeatSecond + time.Duration(rand.Intn(int(maxSrvHeartbeatSecond-minSrvHeartbeatSecond))))
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
		if allRoomsCount, err = s.RenewOnline(context.Background(), s.serverID, roomCount); err != nil {
			time.Sleep(time.Duration(s.c.OnlineTick))
			continue
		}
		for _, bucket := range s.buckets {
			bucket.UpRoomsCount(allRoomsCount)
		}
		time.Sleep(time.Duration(s.c.OnlineTick))
	}
}

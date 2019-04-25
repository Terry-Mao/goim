package comet

import (
	"context"
	"math/rand"
	"time"

	logic "github.com/Terry-Mao/goim/api/logic/grpc"
	"github.com/Terry-Mao/goim/internal/comet/conf"
	log "github.com/golang/glog"
	"github.com/zhenjl/cityhash"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/keepalive"
)

const (
	// 通知logic Refresh client連線狀態最小心跳時間
	minServerHeartbeat = time.Minute * 10

	// 通知logic Refresh client連線狀態最大心跳時間
	maxServerHeartbeat = time.Minute * 30

	// grpc htt2 相關參數
	grpcInitialWindowSize     = 1 << 24
	grpcInitialConnWindowSize = 1 << 24
	grpcMaxSendMsgSize        = 1 << 24
	grpcMaxCallMsgSize        = 1 << 24

	// 心跳包的頻率
	grpcKeepAliveTime = time.Second * 10

	// 心跳回覆如果超過此時間則close連線
	grpcKeepAliveTimeout = time.Second * 3

	// 連線失敗後等待多久才又開始嘗試練線
	grpcBackoffMaxDelay = time.Second * 3
)

func newLogicClient(c *conf.RPCClient) logic.LogicClient {
	// grpc 連線的timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.Dial))
	defer cancel()
	conn, err := grpc.DialContext(ctx, "discovery://default/goim.logic",
		[]grpc.DialOption{
			// 與server溝通不用檢查憑證之類
			grpc.WithInsecure(),

			// Http2相關參數設定
			grpc.WithInitialWindowSize(grpcInitialWindowSize),
			grpc.WithInitialConnWindowSize(grpcInitialConnWindowSize),
			grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(grpcMaxCallMsgSize)),
			grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(grpcMaxSendMsgSize)),

			//
			grpc.WithBackoffMaxDelay(grpcBackoffMaxDelay),

			// 心跳機制參數
			grpc.WithKeepaliveParams(keepalive.ClientParameters{
				Time:                grpcKeepAliveTime,
				Timeout:             grpcKeepAliveTimeout,
				PermitWithoutStream: true,
			}),

			// 設定grpc Load Balancing，需要有service discovery
			grpc.WithBalancerName(roundrobin.Name),
		}...)
	if err != nil {
		panic(err)
	}
	return logic.NewLogicClient(conn)
}

// Server is comet server.
type Server struct {
	c *conf.Config

	// 控管Reader and Writer Buffer 與Timer的Pool
	round *Round

	// 管理buckets，各紀錄部分的Channel與Room
	buckets []*Bucket

	// buckets總數
	bucketIdx uint32

	//
	serverID string

	// Logic service grpc client
	rpcClient logic.LogicClient
}

// NewServer returns a new Server.
func NewServer(c *conf.Config) *Server {
	s := &Server{
		c:         c,
		round:     NewRound(c),
		rpcClient: newLogicClient(c.RPCClient),
	}

	// 初始化Bucket
	s.buckets = make([]*Bucket, c.Bucket.Size)
	s.bucketIdx = uint32(c.Bucket.Size)
	for i := 0; i < c.Bucket.Size; i++ {
		s.buckets[i] = NewBucket(c.Bucket)
	}

	s.serverID = c.Env.Host

	// 統計線上各房間人數
	go s.onlineproc()
	return s
}

// 所有buckets
func (s *Server) Buckets() []*Bucket {
	return s.buckets
}

// 根據user key 採用CityHash32算法除於bucket總數的出來的餘數，來取出某個bucket
// 用意在同時間針對不同房間做推播時可以避免使用到同一把鎖，降低鎖的競爭
// 所以可以調高bucket來增加併發量，但同時會多佔用內存
func (s *Server) Bucket(subKey string) *Bucket {
	idx := cityhash.CityHash32([]byte(subKey), uint32(len(subKey))) % s.bucketIdx
	if conf.Conf.Debug {
		log.Infof("%s hit channel bucket index: %d use cityhash", subKey, idx)
	}
	return s.buckets[idx]
}

// 通知logic Refresh client連線狀態的時間(心跳包的週期)
func (s *Server) RandServerHearbeat() time.Duration {
	return (minServerHeartbeat + time.Duration(rand.Int63n(int64(maxServerHeartbeat-minServerHeartbeat))))
}

func (s *Server) Close() (err error) {
	return
}

// 統計各房間人數並發給logic service做更新
// 因為logic有提供http獲得某房間人數
func (s *Server) onlineproc() {
	for {
		var (
			allRoomsCount map[string]int32
			err           error
		)
		roomCount := make(map[string]int32)

		// 因為房間會分散在不同的bucket所以需要統計
		for _, bucket := range s.buckets {
			for roomID, count := range bucket.RoomsCount() {
				roomCount[roomID] += count
			}
		}
		if allRoomsCount, err = s.RenewOnline(context.Background(), s.serverID, roomCount); err != nil {
			time.Sleep(time.Second)
			continue
		}
		for _, bucket := range s.buckets {
			bucket.UpRoomsCount(allRoomsCount)
		}

		// 每10秒統計一次發給logic
		time.Sleep(time.Second * 10)
	}
}

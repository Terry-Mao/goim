package main

import (
	"net/rpc"
	"strconv"
	"time"

	log "code.google.com/p/log4go"
	"github.com/Terry-Mao/goim/libs/hash/ketama"
	inet "github.com/Terry-Mao/goim/libs/net"
	rproto "github.com/thinkboy/goim/libs/proto/router"
)

var (
	routerServiceMap = map[string]**rpc.Client{}
	routerRing       *ketama.HashRing
)

const (
	routerService               = "RouterRPC"
	routerServicePing           = "RouterRPC.Ping"
	routerServicePut            = "RouterRPC.Put"
	routerServiceDel            = "RouterRPC.Del"
	routerServiceDelServer      = "RouterRPC.DelServer"
	routerServiceAllRoomCount   = "RouterRPC.AllRoomCount"
	routerServiceAllServerCount = "RouterRPC.AllServerCount"
	routerServiceGet            = "RouterRPC.Get"
	routerServiceMGet           = "RouterRPC.MGet"
	routerServiceGetAll         = "RouterRPC.GetAll"
)

func InitRouter() (err error) {
	var (
		network, addr string
	)
	routerRing = ketama.NewRing(ketama.Base)
	for serverId, addrs := range Conf.RouterRPCAddrs {
		// WARN r must every recycle changed for reconnect
		var (
			r          *rpc.Client
			routerQuit = make(chan struct{}, 1)
		)
		if network, addr, err = inet.ParseNetwork(addrs); err != nil {
			log.Error("inet.ParseNetwork() error(%v)", err)
			return
		}
		r, err = rpc.Dial(network, addr)
		if err != nil {
			log.Error("rpc.Dial(\"%s\", \"%s\") error(%v)", network, addr, err)
		}
		go Reconnect(&r, routerQuit, network, addr)
		log.Debug("router rpc addr:%s connect", addr)
		routerServiceMap[serverId] = &r
		routerRing.AddNode(serverId, 1)
	}
	routerRing.Bake()
	return
}

// Reconnect for ping rpc server and reconnect with it when it's crash.
func Reconnect(dst **rpc.Client, quit chan struct{}, network, address string) {
	var (
		tmp    *rpc.Client
		err    error
		call   *rpc.Call
		ch     = make(chan *rpc.Call, 1)
		client = *dst
	)
	for {
		select {
		case <-quit:
			return
		default:
			if client != nil {
				call = <-client.Go(routerServicePing, 0, 0, ch).Done
			}
			if client == nil || call.Error != nil {
				if tmp, err = rpc.Dial(network, address); err == nil {
					*dst = tmp
					client = tmp
				}
			}
		}
		time.Sleep(1 * time.Second)
	}
}

func getRouters() map[string]**rpc.Client {
	return routerServiceMap
}

func getRouterByServer(server string) (*rpc.Client, error) {
	if client, ok := routerServiceMap[server]; !ok || *client == nil {
		return nil, ErrRouter
	} else {
		return *client, nil
	}
}

func getRouterByUID(userID int64) (*rpc.Client, error) {
	return getRouterByServer(routerRing.Hash(strconv.FormatInt(userID, 10)))
}

func getRouterNode(userID int64) string {
	return routerRing.Hash(strconv.FormatInt(userID, 10))
}

func connect(userID int64, server, roomId int32) (seq int32, err error) {
	var client *rpc.Client
	if client, err = getRouterByUID(userID); err != nil {
		return
	}
	arg := &rproto.PutArg{UserId: userID, Server: server, RoomId: roomId}
	reply := &rproto.PutReply{}
	if err = client.Call(routerServicePut, arg, reply); err != nil {
		log.Error("c.Call(\"%s\",\"%v\") error(%v)", routerServicePut, arg, err)
	} else {
		seq = reply.Seq
	}
	return
}

func disconnect(userID int64, seq, roomId int32) (has bool, err error) {
	var (
		client *rpc.Client
		arg    = &rproto.DelArg{UserId: userID, Seq: seq, RoomId: roomId}
		reply  = &rproto.DelReply{}
	)
	if client, err = getRouterByUID(userID); err != nil {
		return
	}
	if err = client.Call(routerServiceDel, arg, reply); err != nil {
		log.Error("c.Call(\"%s\",\"%v\") error(%v)", routerServiceDel, *arg, err)
	} else {
		has = reply.Has
	}
	return
}

func delServer(server int32) (err error) {
	var (
		client **rpc.Client
		arg    = &rproto.DelServerArg{Server: server}
	)
	for _, client = range routerServiceMap {
		if err = (*client).Call(routerServiceDelServer, arg, 0); err != nil {
			log.Error("c.Call(\"%s\",\"%v\") error(%v)", routerServiceDelServer, *arg, err)
		}
	}
	return
}

func allRoomCount(client *rpc.Client) (counter map[int32]int32, err error) {
	var (
		reply = &rproto.AllRoomCountReply{}
	)
	if err = client.Call(routerServiceAllRoomCount, 0, reply); err != nil {
		log.Error("c.Call(\"%s\", nil) error(%v)", routerServiceAllRoomCount, err)
	} else {
		counter = reply.Counter
	}
	return
}

func allServerCount(client *rpc.Client) (counter map[int32]int32, err error) {
	var (
		reply = &rproto.AllServerCountReply{}
	)
	if err = client.Call(routerServiceAllServerCount, 0, reply); err != nil {
		log.Error("c.Call(\"%s\", nil) error(%v)", routerServiceAllServerCount, err)
	} else {
		counter = reply.Counter
	}
	return
}

func genSubKey(userId int64) (res map[int32][]string) {
	var (
		i      int
		ok     bool
		key    string
		keys   []string
		client *rpc.Client
		err    error
		arg    = &rproto.GetArg{UserId: userId}
		reply  = &rproto.GetReply{}
	)
	res = make(map[int32][]string)
	if client, err = getRouterByUID(userId); err != nil {
		return
	}
	if err = client.Call(routerServiceGet, arg, reply); err != nil {
		log.Error("client.Call(\"%s\",\"%v\") error(%v)", routerServiceGet, arg, err)
		return
	}
	for i = 0; i < len(reply.Servers); i++ {
		key = encode(userId, reply.Seqs[i])
		if keys, ok = res[reply.Servers[i]]; !ok {
			keys = []string{}
		}
		keys = append(keys, key)
		res[reply.Servers[i]] = keys
	}
	return
}

func getSubKeys(res chan *rproto.MGetReply, serverId string, userIds []int64) {
	var reply *rproto.MGetReply
	if client, err := getRouterByServer(serverId); err == nil {
		arg := &rproto.MGetArg{UserIds: userIds}
		reply = &rproto.MGetReply{}
		if err = client.Call(routerServiceMGet, arg, reply); err != nil {
			log.Error("client.Call(\"%s\",\"%v\") error(%v)", routerServiceMGet, arg, err)
			reply = nil
		}
	}
	res <- reply
}

func genSubKeys(userIds []int64) (divide map[int32][]string) {
	var (
		i, j, k      int
		node, subkey string
		subkeys      []string
		server       int32
		session      *rproto.GetReply
		reply        *rproto.MGetReply
		uid          int64
		ids          []int64
		ok           bool
		m            = make(map[string][]int64)
		res          = make(chan *rproto.MGetReply, 1)
	)
	divide = make(map[int32][]string) //map[comet.serverId][]subkey
	for i = 0; i < len(userIds); i++ {
		node = getRouterNode(userIds[i])
		if ids, ok = m[node]; !ok {
			ids = []int64{}
		}
		ids = append(ids, userIds[i])
		m[node] = ids
	}
	for node, ids = range m {
		go getSubKeys(res, node, ids)
	}
	k = len(m)
	for k > 0 {
		k--
		if reply = <-res; reply == nil {
			continue
		}
		for j = 0; j < len(reply.UserIds); j++ {
			session = reply.Sessions[j]
			uid = reply.UserIds[j]
			for i = 0; i < len(session.Seqs); i++ {
				subkey = encode(uid, session.Seqs[i])
				server = session.Servers[i]
				if subkeys, ok = divide[server]; !ok {
					subkeys = []string{subkey}
				} else {
					subkeys = append(subkeys, subkey)
				}
				divide[server] = subkeys
			}
		}
	}
	return
}

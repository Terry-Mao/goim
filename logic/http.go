package main

import (
	log "code.google.com/p/log4go"
	"encoding/json"
	"github.com/Terry-Mao/goim/define"
	inet "github.com/Terry-Mao/goim/libs/net"
	rproto "github.com/Terry-Mao/goim/proto/router"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

func InitHTTP() (err error) {
	// http listen
	var network, addr string
	for i := 0; i < len(Conf.HTTPAddrs); i++ {
		httpServeMux := http.NewServeMux()
		httpServeMux.HandleFunc("/1/pushs", Pushs)
		httpServeMux.HandleFunc("/1/push/all", PushAll)
		log.Info("start http listen:\"%s\"", Conf.HTTPAddrs[i])
		if network, addr, err = inet.ParseNetwork(Conf.HTTPAddrs[i]); err != nil {
			log.Error("inet.ParseNetwork() error(%v)", err)
			return
		}
		go httpListen(httpServeMux, network, addr)
	}
	return
}

func httpListen(mux *http.ServeMux, network, addr string) {
	httpServer := &http.Server{Handler: mux, ReadTimeout: Conf.HTTPReadTimeout, WriteTimeout: Conf.HTTPWriteTimeout}
	httpServer.SetKeepAlivesEnabled(true)
	l, err := net.Listen(network, addr)
	if err != nil {
		log.Error("net.Listen(\"%s\", \"%s\") error(%v)", network, addr, err)
		panic(err)
	}
	if err := httpServer.Serve(l); err != nil {
		log.Error("server.Serve() error(%v)", err)
		panic(err)
	}
}

// retPWrite marshal the result and write to client(post).
func retPWrite(w http.ResponseWriter, r *http.Request, res map[string]interface{}, body *string, start time.Time) {
	data, err := json.Marshal(res)
	if err != nil {
		log.Error("json.Marshal(\"%v\") error(%v)", res, err)
		return
	}
	dataStr := string(data)
	if _, err := w.Write([]byte(dataStr)); err != nil {
		log.Error("w.Write(\"%s\") error(%v)", dataStr, err)
	}
	log.Info("req: \"%s\", post: \"%s\", res:\"%s\", ip:\"%s\", time:\"%fs\"", r.URL.String(), *body, dataStr, r.RemoteAddr, time.Now().Sub(start).Seconds())
}

type pushsBodyMsg struct {
	Msg     json.RawMessage `json:"m"`
	UserIds []int64         `json:"u"`
}

func parsePushsBody(body []byte) (msg []byte, userIds []int64, err error) {
	tmp := pushsBodyMsg{}
	if err = json.Unmarshal(body, &tmp); err != nil {
		return
	}
	msg = tmp.Msg
	userIds = tmp.UserIds
	return
}

// {"m":{"test":1},"u":"1,2,3"}
func Pushs(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", 405)
		return
	}
	body := ""
	res := map[string]interface{}{"ret": OK}
	defer retPWrite(w, r, res, &body, time.Now())
	// param
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		res["ret"] = InternalErr
		log.Error("ioutil.ReadAll() failed (%s)", err)
		return
	}
	body = string(bodyBytes)
	log.Debug("pushs msg:%s", body)
	msg, userIds, err := parsePushsBody(bodyBytes)
	if err != nil {
		res["ret"] = InternalErr
		log.Error("parsePushsBody(\"%s\") error(%s)", bodyBytes, err)
		return
	}

	m := divideToRouter(userIds)    //m: map[router.serverId][]userId
	divide, err := divideToComet(m) //divide: map[comet.serverId][]subkey
	if err != nil {
		log.Error("divideToComet() error(%v)", err)
		return
	}
	if len(divide) == 0 {
		log.Debug("no online users")
		res["ret"] = OK
		return
	}

	var (
		cometIds = make([]int32, 0, len(divide))
		subkeys  = make([][]string, 0, len(divide))
	)
	for cometId, keys := range divide {
		cometIds = append(cometIds, cometId)
		subkeys = append(subkeys, keys)
	}

	if err := multiPushTokafka(cometIds, subkeys, msg); err != nil {
		res["ret"] = InternalErr
		log.Error("pushsTokafka(\"%s\") error(%s)", msg, err)
		return
	}

	res["ret"] = OK
	return
}

// get subkeys from all routers and divide by comet-server-id
func divideToComet(m map[string][]int64) (divide map[int32][]string, err error) {
	divide = make(map[int32][]string) //map[comet.serverId][]subkey
	for routerId, us := range m {
		// TODO muti-routine get
		var reply *rproto.MGetReply
		reply, err = getSubkeys(routerId, us)
		if err != nil {
			log.Error("getSubkeys(\"%s\") error(%s)", routerId, err)
			return
		}
		for j := 0; j < len(reply.UserIds); j++ {
			s := reply.Sessions[j]
			log.Debug("sessions seqs:%v serverids:%v", s.Seqs, s.Servers)
			for i := 0; i < len(s.Seqs); i++ {
				subkey := define.Encode(reply.UserIds[j], s.Seqs[i])
				subkeys, ok := divide[s.Servers[i]]
				if !ok {
					subkeys = make([]string, 0, 1000) // TODO:consider
				}
				divide[s.Servers[i]] = append(subkeys, subkey)
			}
		}
	}
	return
}

func PushAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", 405)
		return
	}
	body := ""
	res := map[string]interface{}{"ret": OK}
	defer retPWrite(w, r, res, &body, time.Now())
	// param
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		res["ret"] = InternalErr
		log.Error("ioutil.ReadAll() failed (%v)", err)
		return
	}
	body = string(bodyBytes)
	if err := broadcastTokafka(bodyBytes); err != nil {
		res["ret"] = InternalErr
		log.Error("broadcastTokafka(\"%s\") error(%s)", bodyBytes, err)
		return
	}
	res["ret"] = OK
	return
}

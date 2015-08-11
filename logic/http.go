package main

import (
	log "code.google.com/p/log4go"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

func InitHTTP() error {
	// http listen
	for i := 0; i < len(Conf.HTTPAddrs); i++ {
		httpServeMux := http.NewServeMux()
		//httpServeMux.HandleFunc("/1/push", Push)
		httpServeMux.HandleFunc("/1/pushs", Pushs)
		httpServeMux.HandleFunc("/1/push/all", PushAll)
		log.Info("start http listen:\"%s\"", Conf.HTTPAddrs[i])
		go httpListen(httpServeMux, Conf.HTTPNetworks[i], Conf.HTTPAddrs[i])
	}
	return nil
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

/*
func Push(w http.ResponseWriter, r *http.Request) {
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
	params := r.URL.Query()
	userId, err := strconv.ParseInt(params.Get("userid"), 10, 64)
	if err != nil {
		res["ret"] = InternalErr
		return
	}
	// push to queue
	if err := pushTokafka(userId, bodyBytes); err != nil {
		res["ret"] = InternalErr
		log.Error("pushTokafka(%d) error(%v)", userId, err)
		return
	}
	res["ret"] = OK
	return
}
*/

// {"m":"{"test":1}","u":"1,2,3"}
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

	m := divideNode(userIds)
	divide := make(map[int32][]string) //map[comet.serverId]userIds
	for addr, us := range m {
		reply, err := getSubkeys(addr, us)
		if err != nil {
			res["ret"] = InternalErr
			log.Error("getSubkeys(\"%s\") error(%s)", addr, err)
			return
		}
		log.Debug("getSubkeys:%v addr:%s", reply.UserIds, addr)
		for j := 0; j < len(reply.UserIds); j++ {
			s := reply.Sessions[j]
			log.Debug("sessions seqs:%v serverids:%v", s.Seqs, s.Servers)
			for i := 0; i < len(s.Seqs); i++ {
				subkey := fmt.Sprintf("%d_%d", reply.UserIds[j], s.Seqs[i])
				subkeys, ok := divide[s.Servers[i]]
				if !ok {
					subkeys = make([]string, 0, 1000) //TODO:consider
				}
				divide[s.Servers[i]] = append(subkeys, subkey)
			}
		}
	}

	for cometId, subkeys := range divide {
		if err := pushsTokafka(cometId, subkeys, msg); err != nil {
			res["ret"] = InternalErr
			log.Error("pushsTokafka(cometId:\"%d\") error(%s)", cometId, err)
			return
		}
	}

	res["ret"] = OK
	return
}

type pushsBodyMsg struct {
	Msg     []byte  `json:"m"`
	UserIds []int64 `json:"u"`
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
	log.Debug("pushall msg:%s", body)

	divide := make(map[int32][]string) //map[comet.serverId]userIds
	routers := getRouters()
	for addr, _ := range routers {
		//TODO: muti-routine get
		reply, err := getAllSubkeys(addr)
		if err != nil {
			res["ret"] = InternalErr
			log.Error("getAllSubkeys(\"%s\") error(%s)", addr, err)
			return
		}
		log.Debug("addr:%s getSubkeys:%v", addr, reply.UserIds)
		for j := 0; j < len(reply.UserIds); j++ {
			s := reply.Sessions[j]
			log.Debug("sessions seqs:%v serverids:%v", s.Seqs, s.Servers)
			for i := 0; i < len(s.Seqs); i++ {
				subkey := fmt.Sprintf("%d_%d", reply.UserIds[j], s.Seqs[i])
				subkeys, ok := divide[s.Servers[i]]
				if !ok {
					subkeys = make([]string, 0, 1000) //TODO:consider
				}
				divide[s.Servers[i]] = append(subkeys, subkey)
			}
		}
	}

	for cometId, subkeys := range divide {
		if err := pushsTokafka(cometId, subkeys, bodyBytes); err != nil {
			res["ret"] = InternalErr
			log.Error("pushsTokafka(cometId:\"%d\") error(%s)", cometId, err)
			return
		}
	}

	res["ret"] = OK
	return
}

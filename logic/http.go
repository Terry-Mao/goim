package main

import (
	log "code.google.com/p/log4go"
	"encoding/json"
	inet "github.com/Terry-Mao/goim/libs/net"
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
		httpServeMux.HandleFunc("/1/push", Push)
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

type pushBodyMsg struct {
	Msg    json.RawMessage `json:"m"`
	UserId int64           `json:"u"`
}

func parsePushBody(body []byte) (msg []byte, userId int64, err error) {
	tmp := pushBodyMsg{}
	if err = json.Unmarshal(body, &tmp); err != nil {
		return
	}
	msg = tmp.Msg
	userId = tmp.UserId
	return
}

// {"m":{"test":1},"u":"1,2,3"}
func Push(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", 405)
		return
	}
	var (
		body      string
		server    int32
		keys      []string
		subKeys   map[int32][]string
		bodyBytes []byte
		userId    int64
		err       error
		res       = map[string]interface{}{"ret": OK}
	)
	defer retPWrite(w, r, res, &body, time.Now())
	if bodyBytes, err = ioutil.ReadAll(r.Body); err != nil {
		log.Error("ioutil.ReadAll() failed (%s)", err)
		res["ret"] = InternalErr
		return
	}
	body = string(bodyBytes)
	if bodyBytes, userId, err = parsePushBody(bodyBytes); err != nil {
		log.Error("parsePushsBody(\"%s\") error(%s)", body, err)
		res["ret"] = InternalErr
		return
	}
	subKeys = genSubKey(userId)
	for server, keys = range subKeys {
		if err = mpushKafka(server, keys, bodyBytes); err != nil {
			res["ret"] = InternalErr
			return
		}
	}
	res["ret"] = OK
	return
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
	var (
		body      string
		bodyBytes []byte
		server    int32
		userIds   []int64
		err       error
		res       = map[string]interface{}{"ret": OK}
		subKeys   map[int32][]string
		keys      []string
	)
	defer retPWrite(w, r, res, &body, time.Now())
	if bodyBytes, err = ioutil.ReadAll(r.Body); err != nil {
		log.Error("ioutil.ReadAll() failed (%s)", err)
		res["ret"] = InternalErr
		return
	}
	body = string(bodyBytes)
	if bodyBytes, userIds, err = parsePushsBody(bodyBytes); err != nil {
		log.Error("parsePushsBody(\"%s\") error(%s)", body, err)
		res["ret"] = InternalErr
		return
	}
	subKeys = genSubKeys(userIds)
	for server, keys = range subKeys {
		if err = mpushKafka(server, keys, bodyBytes); err != nil {
			res["ret"] = InternalErr
			return
		}
	}
	res["ret"] = OK
	return
}

func PushAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", 405)
		return
	}
	var (
		bodyBytes []byte
		body      string
		err       error
		ret       = OK
		res       = map[string]interface{}{"ret": ret}
	)
	defer retPWrite(w, r, res, &body, time.Now())
	if bodyBytes, err = ioutil.ReadAll(r.Body); err != nil {
		log.Error("ioutil.ReadAll() failed (%v)", err)
		ret = InternalErr
	} else {
		body = string(bodyBytes)
		if err := broadcastKafka(bodyBytes); err != nil {
			log.Error("broadcastKafka(\"%s\") error(%s)", body, err)
			ret = InternalErr
		}
	}
	res["ret"] = ret
	return
}

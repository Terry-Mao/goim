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

func InitHTTPPush() (err error) {
	var network, addr string
	for i := 0; i < len(Conf.HTTPPushAddrs); i++ {
		httpServeMux := http.NewServeMux()
		httpServeMux.HandleFunc("/1/push", Push)
		log.Info("start http push listen:\"%s\"", Conf.HTTPPushAddrs[i])
		if network, addr, err = inet.ParseNetwork(Conf.HTTPPushAddrs[i]); err != nil {
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

// Push push a message to a specified sub key, must goroutine safe.
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
	params := r.URL.Query()
	key := params.Get("key")
	log.Debug("push key: \"%s\"", key)
	bucket := DefaultServer.Bucket(key)
	if channel := bucket.Get(key); channel != nil {
		// padding let caller do
		if err = channel.PushMsg(1, OP_SEND_SMS_REPLY, bodyBytes); err != nil {
			res["ret"] = InternalErr
			log.Error("channel.PushMsg() error(%v)", err)
			return
		}
	}
	res["ret"] = OK
	return
}

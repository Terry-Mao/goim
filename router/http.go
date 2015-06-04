package main

import (
	log "code.google.com/p/log4go"
	"encoding/json"
	"net"
	"net/http"
	"time"
)

func InitHttp() error {
	// http listen
	for _, bind := range Conf.HttpBind {
		httpServeMux := http.NewServeMux()
		httpServeMux.HandleFunc("/1/sub", sub)
		httpServeMux.HandleFunc("/1/topic", topic)
		log.Info("start http listen addr:\"%s\"", bind)
		go httpListen(httpServeMux, bind)
	}
	return nil
}

func httpListen(mux *http.ServeMux, bind string) {
	httpServer := &http.Server{Handler: mux, ReadTimeout: Conf.HttpReadTimeout, WriteTimeout: Conf.HttpWriteTimeout}
	httpServer.SetKeepAlivesEnabled(true)
	l, err := net.Listen("tcp", bind)
	if err != nil {
		log.Error("net.Listen(\"tcp\", \"%s\") error(%v)", bind, err)
		panic(err)
	}
	if err := httpServer.Serve(l); err != nil {
		log.Error("server.Serve() error(%v)", err)
		panic(err)
	}
}

// retWriter is a json writer for http get method.
func retWriter(r *http.Request, wr http.ResponseWriter, start time.Time, res map[string]interface{}) {
	wr.Header().Set("Content-Type", "application/json;charset=utf-8")
	ret := res["ret"].(int)
	byteJson, err := json.Marshal(res)
	if err != nil {
		log.Error("json.Marshal(\"%v\") failed (%v)", res, err)
		return
	}
	if _, err := wr.Write(byteJson); err != nil {
		log.Error("wr.Write(\"%s\") failed (%v)", string(byteJson), err)
		return
	}
	hs := time.Now().Sub(start)
	if r.Method == "GET" {
		log.Info("[%s]path:%s(params:%s,time:%f,ret:%v)", r.RemoteAddr, r.URL.Path, r.URL.RawQuery, hs.Seconds(), ret)
	} else {
		log.Info("[%s]path:%s(params:%s,time:%f,ret:%v)", r.RemoteAddr, r.URL.Path, r.Form.Encode(), hs.Seconds(), ret)
	}
}

// sub get the info of sub key.
func sub(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	res := map[string]interface{}{"ret": OK}
	defer retWriter(r, w, time.Now(), res)
	params := r.URL.Query()
	key := params.Get("key")
	sb := DefaultBuckets.SubBucket(key)
	if sb == nil {
		log.Error("DefaultBuckets get subbucket error key(%s)", key)
		res["ret"] = InternalErr
		return
	}
	n := sb.Get(key)
	if n == nil {
		res["ret"] = NoExistKey
		return
	}
	data := map[string]interface{}{}
	data["state"] = n.state
	data["server"] = n.server
	res["data"] = data
}

// topic return all sub key in topic.
func topic(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	res := map[string]interface{}{"ret": OK}
	defer retWriter(r, w, time.Now(), res)
	params := r.URL.Query()
	key := params.Get("key")
	tb := DefaultBuckets.TopicBucket(key)
	if tb == nil {
		log.Error("DefaultBuckets get topicbucket error key(%s)", key)
		res["ret"] = InternalErr
		return
	}
	res["data"] = tb.Get(key)
}

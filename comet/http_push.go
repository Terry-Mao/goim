package main

import (
	log "code.google.com/p/log4go"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

func InitHttpPush() error {
	// http listen
	for _, bind := range Conf.HttpPushBind {
		httpServeMux := http.NewServeMux()
		httpServeMux.HandleFunc("/1/push", Push)
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
		if err = channel.Push(1, OP_SEND_SMS_REPLY, bodyBytes); err != nil {
			res["ret"] = InternalErr
			log.Error("channel.Push() error(%v)", err)
			return
		}
	}
	res["ret"] = OK
	return
}

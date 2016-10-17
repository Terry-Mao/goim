package main

import (
	"encoding/json"
	"fmt"
	log "github.com/thinkboy/log4go"
	"net/http"
)

const (
	OK = 1
)

type Monitor struct {
}

// StartPprof start http monitor.
func InitMonitor(binds []string) {
	m := new(Monitor)
	monitorServeMux := http.NewServeMux()
	monitorServeMux.HandleFunc("/monitor/ping", m.Ping)
	monitorServeMux.HandleFunc("/monitor/stat", m.Stat)
	for _, addr := range binds {
		log.Info("start monitor listen: \"%s\"", addr)
		go func(bind string) {
			if err := http.ListenAndServe(bind, monitorServeMux); err != nil {
				log.Error("http.ListenAndServe(\"%s\", pprofServeMux) error(%v)", addr, err)
				panic(err)
			}
		}(addr)
	}
}

// monitor ping
func (m *Monitor) Ping(w http.ResponseWriter, r *http.Request) {
	for _, c := range cometServiceMap {
		if err := c.rpcClient.Available(); err != nil {
			http.Error(w, fmt.Sprintf("ping rpc error(%v)", err), http.StatusInternalServerError)
			return
		}
	}
	w.Write([]byte("ok"))
}

// monitor stat
func (m *Monitor) Stat(w http.ResponseWriter, r *http.Request) {
	var (
		err error
		b   []byte
		res = map[string]interface{}{"ret": OK}
	)
	switch r.Method {
	case "GET":
		res["data"] = DefaultStat.Info()
	case "DELETE":
		DefaultStat.Reset()
	}
	if b, err = json.Marshal(res); err != nil {
		log.Error("json.Marshal(%v) error(%v)", res, err)
		return
	}
	w.Write(b)
}

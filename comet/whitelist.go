package main

import (
	"log"
	"os"
	"strings"
)

type Whitelist struct {
	Log  *log.Logger
	list map[string]struct{} // whitelist for debug
}

// NewWhitelist a whitelist struct.
func NewWhitelist(file string, list []string) (w *Whitelist, err error) {
	var (
		key string
		f   *os.File
	)
	if f, err = os.OpenFile(file, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644); err == nil {
		w = new(Whitelist)
		w.Log = log.New(f, "", log.LstdFlags)
		w.list = make(map[string]struct{})
		for _, key = range list {
			w.list[key] = struct{}{}
		}
	}
	return
}

// Contains whitelist contains a key or not.
func (w *Whitelist) Contains(key string) (ok bool) {
	if ix := strings.Index(key, "_"); ix > -1 {
		_, ok = w.list[key[:ix]]
	}
	return
}

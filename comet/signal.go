package main

import (
	"os"
	"os/signal"
	"syscall"

	log "github.com/thinkboy/log4go"
)

// InitSignal register signals handler.
func InitSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT, syscall.SIGSTOP)
	for {
		s := <-c
		log.Info("comet[%s] get a signal %s", Ver, s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGSTOP, syscall.SIGINT:
			return
		case syscall.SIGHUP:
			reload()
		default:
			return
		}
	}
}

func reload() {
	newConf, err := ReloadConfig()
	if err != nil {
		log.Error("ReloadConfig() error(%v)", err)
		return
	}
	Conf = newConf
}

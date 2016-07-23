package main

import (
	"log"
	"os"
)

var (
	WhiteLog *log.Logger
)

func InitWhiteList(file string) (err error) {
	var f *os.File
	if f, err = os.OpenFile(file, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666); err == nil {
		WhiteLog = log.New(f, "", log.LstdFlags)
	}
	return
}

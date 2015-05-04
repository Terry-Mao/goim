package main

import (
	log "code.google.com/p/log4go"
	"crypto/rsa"
	myrsa "github.com/Terry-Mao/goim/libs/crypto/rsa"
	"io/ioutil"
	"os"
)

var (
	RSAPri *rsa.PrivateKey
)

func InitRSA() (err error) {
	var (
		pri  []byte
		file *os.File
	)
	if file, err = os.Open(Conf.RSAPrivate); err != nil {
		log.Error("os.Open(\"%s\") error(%v)", Conf.RSAPrivate, err)
		return
	}
	if pri, err = ioutil.ReadAll(file); err != nil {
		log.Error("ioutil.ReadAll(\"%s\") error(%v)", Conf.RSAPrivate, err)
		return
	}
	log.Debug("private.pem : \n%s", string(pri))
	if RSAPri, err = myrsa.PrivateKey(pri); err != nil {
		log.Error("myrsa.PrivateKey() error(%v)", err)
	}
	return
}

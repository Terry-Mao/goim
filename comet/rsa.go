package main

import (
	log "code.google.com/p/log4go"
	"crypto/rsa"
	myrsa "github.com/Terry-Mao/goim/libs/crypto/rsa"
	"io/ioutil"
)

var (
	RSAPri *rsa.PrivateKey
)

func InitRSA() (err error) {
	var pri []byte
	if pri, err = ioutil.ReadAll(Conf.RSAPrivate); err != nil {
		log.Errror("ioutil.ReadAll(\"%s\") error(%v)", Conf.RSAPrivate, err)
		return
	}
	if RSAPri, err = myrsa.PrivateKey(pri); err != nil {
		log.Errror("myrsa.PrivateKey() error(%v)", err)
	}
	return
}

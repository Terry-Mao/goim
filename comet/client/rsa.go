package main

import (
	log "code.google.com/p/log4go"
	"crypto/rsa"
	myrsa "github.com/Terry-Mao/goim/libs/crypto/rsa"
	"io/ioutil"
	"os"
)

var (
	RSAPub *rsa.PublicKey
)

func InitRSA() (err error) {
	var (
		pub  []byte
		file *os.File
	)
	if file, err = os.Open(Conf.RSAPublic); err != nil {
		log.Error("os.Open(\"%s\") error(%v)", Conf.RSAPublic, err)
		return
	}
	if pub, err = ioutil.ReadAll(file); err != nil {
		log.Error("ioutil.ReadAll(\"%s\") error(%v)", Conf.RSAPublic, err)
		return
	}
	log.Debug("public.pem : \n%s", string(pub))
	if RSAPub, err = myrsa.PublicKey(pub); err != nil {
		log.Error("myrsa.Public() error(%v)", err)
	}
	return
}

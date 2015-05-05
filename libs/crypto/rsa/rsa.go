package rsa

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

var (
	ErrPrivateKey = errors.New("private key error")
	ErrPublicKey  = errors.New("public key error")
)

func PrivateKey(pri []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pri)
	if block == nil {
		return nil, ErrPrivateKey
	}
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

func PublicKey(pub []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(pub)
	if block == nil {
		return nil, ErrPublicKey
	}
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	rsaPub := pubInterface.(*rsa.PublicKey)
	return rsaPub, nil
}

func Encrypt(orig []byte, pub *rsa.PublicKey) ([]byte, error) {
	return rsa.EncryptPKCS1v15(rand.Reader, pub, orig)
}

func Decrypt(cipher []byte, pri *rsa.PrivateKey) ([]byte, error) {
	return rsa.DecryptPKCS1v15(nil, pri, cipher)
}

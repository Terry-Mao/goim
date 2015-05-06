package rsa

import (
	"testing"
)

const (
	priKey = `
-----BEGIN RSA PRIVATE KEY-----
MFECAQACDQDt0G4B3JeeHjLWvX0CAwEAAQINANmKZncRf2SzCt/qiQIHAP1hu7hC
NwIHAPBFhAcz6wIHAMKsRD3dIQIGDn4S7aBLAgY5OcfnuCQ=
-----END RSA PRIVATE KEY-----
`
	pubKey = `
-----BEGIN PUBLIC KEY-----
MCgwDQYJKoZIhvcNAQEBBQADFwAwFAINAO3QbgHcl54eMta9fQIDAQAB
-----END PUBLIC KEY-----
`
)

func TestRSA(t *testing.T) {
	pri, err := PrivateKey([]byte(priKey))
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	pub, err := PublicKey([]byte(pubKey))
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	msg := "1"
	cipher, err := Encrypt([]byte(msg), pub)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	ori, err := Decrypt(cipher, pri)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if string(ori) != msg {
		t.FailNow()
	}
}

func BenchmarkRSA(b *testing.B) {
	pri, err := PrivateKey([]byte(priKey))
	if err != nil {
		b.Error(err)
		b.FailNow()
	}
	pub, err := PublicKey([]byte(pubKey))
	if err != nil {
		b.Error(err)
		b.FailNow()
	}
	msg := "1"
	cipher, err := Encrypt([]byte(msg), pub)
	if err != nil {
		b.Error(err)
		b.FailNow()
	}
	for i := 0; i < b.N; i++ {
		if _, err := Decrypt(cipher, pri); err != nil {
			b.Error(err)
			b.FailNow()
		}
	}
}

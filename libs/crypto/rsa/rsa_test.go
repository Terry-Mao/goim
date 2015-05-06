package rsa

import (
	"testing"
)

const (
	priKey = `
-----BEGIN RSA PRIVATE KEY-----
MIGTAgEAAhwA3h6/w3C9rE/bZ9C/99QZzz+42d+Md14KCpQ9AgMBAAECG0MCLcHN
MjtYgA1KqY8WXD/pIvxA0OAWRCiA1QIODyI7B8E9NtKx0qExxQ8CDg6tWyaQ7V4+
4YlwMQnzAg4M85efWGJiA8kJYMiuQwIOCPUZ5R6cD2G3CccT1qsCDgN060UY1a0K
xLj6/Fgx
-----END RSA PRIVATE KEY-----
`
	pubKey = `
-----BEGIN PUBLIC KEY-----
MDcwDQYJKoZIhvcNAQEBBQADJgAwIwIcAN4ev8NwvaxP22fQv/fUGc8/uNnfjHde
CgqUPQIDAQAB
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
	msg := "1234567890123456"
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
	msg := "1234567890123456"
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

package rsa

import (
	"testing"
)

const (
	priKey = `
-----BEGIN RSA PRIVATE KEY-----
MIICXQIBAAKBgQC0uoYIqecHK2c9CgyEKWDK5XGrYLT29CgENUm9eBPi4YyCGCXq
aesdRs1TS7X7JKpAh114BGkkNPuEEFHbzIgIHSoNGIB9r/ustGGggdeqqFiEhq6v
xWM85RPWBGxv3WNAnwVqZ+NJ5+1Q0Rwpaazr6wr6LddByFzf/U88GQfzhQIDAQAB
AoGBALE6qO4eD1zMh3UoQZXpLe5KiunQ8CWs0QEvcJzJAFdhb/Sz0ZrLO7F+GSQx
/sfF8N9O364uRR0oh+2+Q0gUjuAgE1dUvzQbQqaygzrs8JiElFQtun9LpUUd9SyI
1jjUhY2/VYW+wKMurUm9DM6bWsyvVkLve1IUCUBEoXRd/OOBAkEA3P5FUHfpmfoq
IHfdfE1d1tZHH5QC0hBxzC4qdsZcg8r3bGssSgSJEz4ujM23kiVIZ5QrdmqeBrq9
G5G5a+6PrQJBANFbcG9OkejCpT50jQXomnTingOTA/xGkcnPqAYoNkPJ5KOi16H6
uXDDv/cRU3TdiO2lzTY6daXla9lF2PkYjjkCQQDCm2eOpQohfhr63JM+kyK/vZKE
TGLveWu80iqyzZtKs8GOyBIIXFYZi/iSJdYx7IMGM4TSkrD2XBuL25fdZAdBAkBm
qHfRnK1ffVKZ9XzRUOWsOxNQnV5u7gu+8dxqaH1zcCR1OPyTqOYVrWcMN6q8u4TR
Q2QFG1VlK8JeoClsu+XBAkAsyXoZ7X6FGclZawtmSgU6Rxv/bBHyRQNhIUxhOTXQ
03ZzpyZ2hrnA4sCQp+nNgzinzttlMub3JEa3KNGzVFys
-----END RSA PRIVATE KEY-----
    `
	pubKey = `
-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQC0uoYIqecHK2c9CgyEKWDK5XGr
YLT29CgENUm9eBPi4YyCGCXqaesdRs1TS7X7JKpAh114BGkkNPuEEFHbzIgIHSoN
GIB9r/ustGGggdeqqFiEhq6vxWM85RPWBGxv3WNAnwVqZ+NJ5+1Q0Rwpaazr6wr6
LddByFzf/U88GQfzhQIDAQAB
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
	msg := "woaini"
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

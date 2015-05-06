package aes

import (
	"encoding/hex"
	"fmt"
	"github.com/Terry-Mao/goim/libs/crypto/padding"
	"testing"
)

func TestAes(t *testing.T) {
	a := []byte("1111111111111111")
	b, err := ECBEncrypt(a, a, padding.PKCS5)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("%s\n", hex.EncodeToString(a))
	fmt.Printf("%s\n", hex.EncodeToString(b))
	c, err := ECBDecrypt(b, a, padding.PKCS5)
	if err != nil {
		t.Error(err)
	}
	if string(a) != string(c) {
		t.Error("decrypt error")
	}
	b, err = CBCEncrypt(a, a, a, padding.PKCS5)
	if err != nil {
		t.Error(err)
	}
	c, err = CBCDecrypt(b, a, a, padding.PKCS5)
	if err != nil {
		t.Error(err)
	}
	if string(a) != string(c) {
		t.Error("decrypt error")
	}
	d := []byte("1")
	b, err = ECBEncrypt(d, a, padding.PKCS5)
	if err != nil {
		t.Error(err)
	}
	c, err = ECBDecrypt(b, a, padding.PKCS5)
	if err != nil {
		t.Error(err)
	}
	if string(d) != string(c) {
		t.Error("decrypt error")
	}
	b, err = CBCEncrypt(d, a, a, padding.PKCS5)
	if err != nil {
		t.Error(err)
	}
	c, err = CBCDecrypt(b, a, a, padding.PKCS5)
	if err != nil {
		t.Error(err)
	}
	if string(d) != string(c) {
		t.Error("decrypt error")
	}
}

func BenchmarkAES(b *testing.B) {
	a := []byte("1111111111111111")
	d, err := ECBEncrypt(a, a, padding.PKCS5)
	if err != nil {
		b.Error(err)
		b.FailNow()
	}
	for i := 0; i < b.N; i++ {
		_, err := ECBDecrypt(d, a, padding.PKCS5)
		if err != nil {
			b.Error(err)
			b.FailNow()
		}
	}
}

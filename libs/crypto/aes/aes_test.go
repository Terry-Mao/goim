package aes

import (
	"crypto/aes"
	"encoding/hex"
	"fmt"
	"testing"
)

func TestAes(t *testing.T) {
	a := []byte("1111111111111111")
	block, err := aes.NewCipher(a)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	b, err := ECBEncrypt(block, a)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("%s\n", hex.EncodeToString(a))
	fmt.Printf("%s\n", hex.EncodeToString(b))
	if string(a) != string(b) {
		t.FailNow()
	}
}

/*
func BenchmarkAES(b *testing.B) {
	a := []byte("1111111111111111")
	o := make([]byte, 50)
	d, err := ECBEncrypt(a, o, a, padding.PKCS5)
	if err != nil {
		b.Error(err)
		b.FailNow()
	}
	for i := 0; i < b.N; i++ {
		_, err := ECBDecrypt(d, o, a, padding.PKCS5)
		if err != nil {
			b.Error(err)
			b.FailNow()
		}
	}
}
*/

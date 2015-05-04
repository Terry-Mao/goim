package padding

import (
	"testing"
)

func TestPKCS5(t *testing.T) {
	a := []byte{1}
	b := PKCS5.Padding(a, 16)
	// pad 15 length
	for i := 15; i > 0; i-- {
		if int(b[i]) != 15 {
			t.Error("padding error")
		}
	}
	if b[0] != 1 {
		t.Error("padding error")
	}
	c, err := PKCS5.Unpadding(b, 16)
	if err != nil {
		t.Error(err)
	}
	if len(c) != 1 {
		t.Error("padding error")
	}
	if c[0] != 1 {
		t.Error("padding error")
	}
}

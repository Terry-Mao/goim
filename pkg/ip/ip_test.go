package ip

import (
	"fmt"
	"testing"
)

func TestIP(t *testing.T) {
	ip := InternalIP()
	fmt.Println(ip)
	if ip == "" {
		t.FailNow()
	}
}

package ioutil

import (
	"bufio"
)

func ReadAll(rd *bufio.Reader, d []byte) (err error) {
	tl, n, t := len(d), 0, 0
	for {
		if t, err = rd.Read(d[n:]); err != nil {
			return
		}
		if n += t; n == tl {
			break
		}
	}
	return
}

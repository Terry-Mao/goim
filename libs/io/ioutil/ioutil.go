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

func ReadBigEndianInt16(rd *bufio.Reader) (d int16, err error) {
	var b [2]byte
	if b[0], err = rd.ReadByte(); err != nil {
		return
	}
	if b[1], err = rd.ReadByte(); err != nil {
		return
	}
	d = int16(b[1]) | int16(b[0])<<8
	return
}

func ReadBigEndianInt32(rd *bufio.Reader) (d int32, err error) {
	var b [4]byte
	if b[0], err = rd.ReadByte(); err != nil {
		return
	}
	if b[1], err = rd.ReadByte(); err != nil {
		return
	}
	if b[2], err = rd.ReadByte(); err != nil {
		return
	}
	if b[3], err = rd.ReadByte(); err != nil {
		return
	}
	d = int32(b[3]) | int32(b[2])<<8 | int32(b[1])<<16 | int32(b[0])<<24
	return
}

func WriteBigEndianInt16(wr *bufio.Writer, v int16) (err error) {
	if err = wr.WriteByte(byte(v >> 8)); err != nil {
		return
	}
	return wr.WriteByte(byte(v))
}

func WriteBigEndianInt32(wr *bufio.Writer, v int32) (err error) {
	if err = wr.WriteByte(byte(v >> 24)); err != nil {
		return
	}
	if err = wr.WriteByte(byte(v >> 16)); err != nil {
		return
	}
	if err = wr.WriteByte(byte(v >> 8)); err != nil {
		return
	}
	return wr.WriteByte(byte(v))
}

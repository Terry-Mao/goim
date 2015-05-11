package aes

import (
	"crypto/cipher"
	"errors"
)

var (
	ErrBlockSize  = errors.New("input not full blocks")
	ErrOutputSize = errors.New("output smaller than input")
)

func encryptBlocks(b cipher.Block, src, dst []byte) error {
	if len(src)%b.BlockSize() != 0 {
		return ErrBlockSize
	}
	if len(dst) < len(src) {
		return ErrOutputSize
	}
	for len(src) > 0 {
		b.Encrypt(dst, src[:b.BlockSize()])
		src = src[b.BlockSize():]
		dst = dst[b.BlockSize():]
	}
	return nil
}

func decryptBlocks(b cipher.Block, dst, src []byte) error {
	if len(src)%b.BlockSize() != 0 {
		return ErrBlockSize
	}
	if len(dst) < len(src) {
		return ErrOutputSize
	}
	for len(src) > 0 {
		b.Decrypt(dst, src[:b.BlockSize()])
		src = src[b.BlockSize():]
		dst = dst[b.BlockSize():]
	}
	return nil
}

func ECBEncrypt(b cipher.Block, src []byte) (dst []byte, err error) {
	// use same buf
	dst = src
	err = encryptBlocks(b, dst, src)
	return
}

func ECBDecrypt(b cipher.Block, src []byte) (dst []byte, err error) {
	// use same buf
	dst = src
	err = decryptBlocks(b, dst, src)
	return
}

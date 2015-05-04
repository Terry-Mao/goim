package padding

import (
	"errors"
)

var (
	ErrPaddingSize = errors.New("padding size error")
)

// Padding is interface used for crypto.
type Padding interface {
	Padding(src []byte, blockSize int) []byte
	Unpadding(src []byte, blockSize int) ([]byte, error)
}

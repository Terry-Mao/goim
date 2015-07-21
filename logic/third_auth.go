package main

// developer could implement "ThirdAuth" interface for decide how get userID
type ThirdAuth interface {
	CheckUID(token string) int64
}

var DefaultThird = NewDefaultThird()

type Third struct {
}

func NewDefaultThird() *Third {
	return &Third{}
}

func (t *Third) CheckUID(token string) (userID int64) {
	return 111
}

package main

// developer could implement "ThirdAuth" interface for decide how get userID
type Auther interface {
	Auth(token string) int64
}

type DefaultAuther struct {
}

func NewDefaultAuther() *DefaultAuther {
	return &DefaultAuther{}
}

func (a *DefaultAuther) Auth(token string) (userID int64) {
	return 0
}

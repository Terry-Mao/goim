package main

import (
	"goim/libs/define"
	"strconv"
)

// developer could implement "Auth" interface for decide how get userId, or roomId
type Auther interface {
	Auth(token string) (userId int64, roomId int32)
}

type DefaultAuther struct {
}

func NewDefaultAuther() *DefaultAuther {
	return &DefaultAuther{}
}

func (a *DefaultAuther) Auth(token string) (userId int64, roomId int32) {

	var err error

	var token0, token1 string
	token0 = strings.Split(token, ",")[0]

	token1 = strings.Split(token, ",")[1]

	if userId, err = strconv.ParseInt(token0, 10, 64); err != nil {
		userId = 0
	}
	var roomIdTemp int64
	if roomIdTemp, err = strconv.ParseInt(token1, 10, 64); err != nil {
		roomId = define.NoRoom
	} else {
		roomId = int32(roomIdTemp)
	}

	return
}

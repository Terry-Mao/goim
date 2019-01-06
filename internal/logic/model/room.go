package model

import (
	"fmt"
	"net/url"
)

// EncodeRoomKey encode a room key.
func EncodeRoomKey(typ string, room string) string {
	return fmt.Sprintf("%s://%s", typ, room)
}

// DecodeRoomKey decode room key.
func DecodeRoomKey(key string) (string, string, error) {
	u, err := url.Parse(key)
	if err != nil {
		return "", "", err
	}
	return u.Scheme, u.Host, nil
}

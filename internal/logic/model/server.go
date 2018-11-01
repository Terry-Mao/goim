package model

// ServerInfo server info.
type ServerInfo struct {
	Server    string   `json:"server"`
	IPAddrs   []string `json:"ip_addrs"`
	IPCount   int32    `json:"ip_count"`
	ConnCount int32    `json:"conn_count"`
	Updated   int64    `json:"updated"`
}

package model

type Presence struct {
	UserId   int32  `json:"user_id"`
	State    string `json:"state"`
	LastSeen int64  `json:"last_seen"`
}

var (
	StateOnline  = "online"
	StateOffline = "offline"
)

package model

import "time"

type IconJob struct {
	Id        int       `json:"id"`
	ProfileId int32     `json:"profile_id"`
	IconUrl   string    `json:"icon_url"`
	CreatedAt time.Time `json:"created_at"`
}

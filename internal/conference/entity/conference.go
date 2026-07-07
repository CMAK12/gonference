package entity

import "time"

type Conference struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	CreatorID      string    `json:"creator_id"`
	InvitedMembers string    `json:"invited_members"`
	Token          string    `json:"token"`
	StartTime      time.Time `json:"start_time"`
	EndTime        time.Time `json:"end_time"`
	CreatedAt      time.Time `json:"created_at"`
}

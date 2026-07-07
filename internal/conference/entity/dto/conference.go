package dto

type CreateConferenceRequestDTO struct {
	Name           string `json:"name"`
	CreatorID      string `json:"creator_id"`
	InvitedMembers string `json:"invited_members"`
	StartTime      string `json:"start_time"`
	EndTime        string `json:"end_time"`
	CreatedAt      string `json:"created_at"`
}

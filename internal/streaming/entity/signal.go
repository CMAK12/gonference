package entity

import (
	"github.com/pion/webrtc/v3"
)

type MessageType string

const (
	TypeOffer     MessageType = "offer"
	TypeAnswer    MessageType = "answer"
	TypeCandidate MessageType = "candidate"
	TypeLeave     MessageType = "leave"
)

type SignalMessage struct {
	Type      MessageType              `json:"type"`
	RoomID    string                   `json:"roomId"`
	MemberID  string                   `json:"memberId"`
	SDP       *string                  `json:"sdp,omitempty"`
	Candidate *webrtc.ICECandidateInit `json:"candidate,omitempty"`
}

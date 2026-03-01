package entity

import (
	"gonference/pkg/pb"

	"github.com/pion/webrtc/v3"
)

type MessageType string

const (
	TypeOffer     MessageType = "offer"
	TypeAnswer    MessageType = "answer"
	TypeCandidate MessageType = "candidate"
	TypeLeave     MessageType = "leave"
)

type Message struct {
	Type      MessageType              `json:"type"`
	RoomID    string                   `json:"roomId"`
	MemberID  string                   `json:"memberId"`
	SDP       *string                  `json:"sdp,omitempty"`
	Candidate *webrtc.ICECandidateInit `json:"candidate,omitempty"`
}

func MapProtoToMessage(msg *pb.SignalMessage) Message {
	return Message{
		Type:     MessageType(msg.Type),
		RoomID:   msg.RoomId,
		MemberID: msg.MemberId,
		SDP:      msg.Sdp,
		Candidate: func() *webrtc.ICECandidateInit {
			if msg.Candidate == nil {
				return nil
			}
			idx := uint16(*msg.Candidate.SdpMlineIndex)
			return &webrtc.ICECandidateInit{
				Candidate:        msg.Candidate.Candidate,
				SDPMid:           msg.Candidate.SdpMid,
				SDPMLineIndex:    &idx,
				UsernameFragment: msg.Candidate.UsernameFragment,
			}
		}(),
	}
}

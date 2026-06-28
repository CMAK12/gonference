package mapper

import (
	impb "github.com/CMAK12/gonference/internal/gen/streaming/v1"
	"github.com/CMAK12/gonference/internal/streaming/entity"
	"github.com/pion/webrtc/v3"
)

func MapProtoToMessage(msg *impb.SignalMessage) entity.SignalMessage {
	return entity.SignalMessage{
		Type:     entity.MessageType(msg.Type),
		RoomID:   msg.RoomId,
		MemberID: msg.MemberId,
		SDP:      msg.Sdp,
		Candidate: func() *webrtc.ICECandidateInit {
			if msg.Candidate == nil {
				return nil
			}
			return &webrtc.ICECandidateInit{
				Candidate:        msg.Candidate.Candidate,
				SDPMid:           msg.Candidate.SdpMid,
				SDPMLineIndex:    new(uint16(*msg.Candidate.SdpMlineIndex)),
				UsernameFragment: msg.Candidate.UsernameFragment,
			}
		}(),
	}
}

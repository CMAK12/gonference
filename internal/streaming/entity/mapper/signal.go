package mapper

import (
	impb "github.com/CMAK12/gonference/internal/gen/streaming/v1"
	"github.com/CMAK12/gonference/internal/streaming/entity"

	"github.com/pion/webrtc/v3"
)

func MapProtoToMessage(msg *impb.SignalMessage) entity.SignalMessage {
	return entity.SignalMessage{
		Type:      protoToType(msg.Type),
		RoomID:    msg.RoomId,
		MemberID:  msg.MemberId,
		SDP:       msg.Sdp,
		Candidate: protoToCandidate(msg.Candidate),
	}
}

func MapMessageToProto(msg entity.SignalMessage) *impb.SignalMessage {
	return &impb.SignalMessage{
		Type:      typeToProto(msg.Type),
		RoomId:    msg.RoomID,
		MemberId:  msg.MemberID,
		Sdp:       msg.SDP,
		Candidate: candidateToProto(msg.Candidate),
	}
}

func protoToType(t impb.MessageType) entity.MessageType {
	switch t {
	case impb.MessageType_ANSWER:
		return entity.TypeAnswer
	case impb.MessageType_CANDIDATE:
		return entity.TypeCandidate
	case impb.MessageType_LEAVE:
		return entity.TypeLeave
	default:
		return entity.TypeOffer
	}
}

func typeToProto(t entity.MessageType) impb.MessageType {
	switch t {
	case entity.TypeAnswer:
		return impb.MessageType_ANSWER
	case entity.TypeCandidate:
		return impb.MessageType_CANDIDATE
	case entity.TypeLeave:
		return impb.MessageType_LEAVE
	default:
		return impb.MessageType_OFFER
	}
}

func protoToCandidate(c *impb.ICECandidate) *webrtc.ICECandidateInit {
	if c == nil {
		return nil
	}

	var mline *uint16
	if c.SdpMlineIndex != nil {
		mline = new(uint16(*c.SdpMlineIndex))
	}

	return &webrtc.ICECandidateInit{
		Candidate:        c.Candidate,
		SDPMid:           c.SdpMid,
		SDPMLineIndex:    mline,
		UsernameFragment: c.UsernameFragment,
	}
}

func candidateToProto(c *webrtc.ICECandidateInit) *impb.ICECandidate {
	if c == nil {
		return nil
	}

	var mline *uint32
	if c.SDPMLineIndex != nil {
		mline = new(uint32(*c.SDPMLineIndex))
	}

	return &impb.ICECandidate{
		Candidate:        c.Candidate,
		SdpMid:           c.SDPMid,
		SdpMlineIndex:    mline,
		UsernameFragment: c.UsernameFragment,
	}
}

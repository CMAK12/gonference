package signaling

import (
	"fmt"

	"github.com/CMAK12/gonference/internal/streaming/entity"
	"github.com/CMAK12/gonference/internal/streaming/sfu"

	"github.com/pion/webrtc/v3"
)

type SFU interface {
	GetOrCreateRoom(id string) *sfu.Room
}

type Signaling struct {
	sfu SFU
}

func New(sfu SFU) *Signaling {
	return &Signaling{
		sfu: sfu,
	}
}

// HandleOffer bootstraps a peer from the client's offer and returns the answer
// to be delivered over the unary RPC. Every subsequent signaling message
// (renegotiation offers/answers, leave) flows over the peer's DataChannel.
func (s *Signaling) HandleOffer(message entity.SignalMessage) (entity.SignalMessage, error) {
	if message.SDP == nil {
		return entity.SignalMessage{}, fmt.Errorf("signaling.HandleOffer: missing offer SDP")
	}

	room := s.sfu.GetOrCreateRoom(message.RoomID)

	offer := webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  *message.SDP,
	}

	answer, _, err := room.AddPeer(offer, message.MemberID)
	if err != nil {
		return entity.SignalMessage{}, fmt.Errorf("signaling.HandleOffer: failed to add peer: %w", err)
	}

	return entity.SignalMessage{
		Type:     entity.TypeAnswer,
		RoomID:   message.RoomID,
		MemberID: message.MemberID,
		SDP:      &answer.SDP,
	}, nil
}

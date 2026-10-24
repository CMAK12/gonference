package signaling

import (
	"context"
	"fmt"

	"github.com/CMAK12/gonference/internal/streaming/entity"
	"github.com/CMAK12/gonference/internal/streaming/sfu"

	"github.com/pion/webrtc/v4"
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

func (s *Signaling) HandleOffer(ctx context.Context, message entity.SignalMessage) (entity.SignalMessage, error) {
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

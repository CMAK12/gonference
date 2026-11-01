package signaling

import (
	"context"
	"fmt"

	"github.com/CMAK12/gonference/internal/streaming/entity"

	"github.com/pion/webrtc/v4"
)

func (s *Signaling) HandleOffer(ctx context.Context, message entity.SignalMessage) (entity.SignalMessage, error) {
	owner, err := s.resolveOwner(ctx, message.RoomID)
	if err != nil {
		return entity.SignalMessage{}, fmt.Errorf("signaling.HandleOffer: %w", err)
	}

	room := s.sfu.GetOrCreateRoom(message.RoomID)

	if owner != s.self {
		if err := s.connectRelay(ctx, room, owner); err != nil {
			return entity.SignalMessage{}, fmt.Errorf("signaling.HandleOffer: connect relay to %s: %w", owner, err)
		}
	}

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

package signaling

import (
	"context"
	"errors"
	"fmt"

	"github.com/CMAK12/gonference/internal/streaming/entity"
	"github.com/CMAK12/gonference/internal/streaming/sfu"
	"github.com/CMAK12/gonference/internal/streaming/sfu/peers"

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

	peer, err := sfu.NewPeer(room, message.MemberID, peers.NewPeer)
	if err != nil {
		return entity.SignalMessage{}, fmt.Errorf("signaling.HandleOffer: failed to create peer: %w", err)
	}

	answer, err := peer.CreateAnswer(offer)
	if err != nil {
		return entity.SignalMessage{}, fmt.Errorf("signaling.HandleOffer: failed to create answer: %w", errors.Join(err, peer.Close()))
	}

	room.Attach(peer)

	return entity.SignalMessage{
		Type:     entity.TypeAnswer,
		RoomID:   message.RoomID,
		MemberID: message.MemberID,
		SDP:      &answer.SDP,
	}, nil
}

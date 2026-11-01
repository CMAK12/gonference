package signaling

import (
	"context"
	"fmt"

	gnpb "github.com/CMAK12/gonference/internal/gen/streaming/v1"
	"github.com/CMAK12/gonference/internal/streaming/entity"
	"github.com/CMAK12/gonference/internal/streaming/sfu"

	"github.com/pion/webrtc/v4"
)

func (s *Signaling) HandleRelayOffer(ctx context.Context, offer entity.RelayOffer) (entity.RelayAnswer, error) {
	room := s.sfu.GetOrCreateRoom(offer.RoomID)

	answer, _, err := room.AddRelayPeer(webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  offer.SDP,
	}, offer.RelayID)
	if err != nil {
		return entity.RelayAnswer{}, fmt.Errorf("signaling.HandleRelayOffer: failed to add relay peer: %w", err)
	}

	return entity.RelayAnswer{
		RoomID:  offer.RoomID,
		RelayID: offer.RelayID,
		SDP:     answer.SDP,
	}, nil
}

func (s *Signaling) HandleRelayLeave(ctx context.Context, roomID, relayID string) {
	room, ok := s.sfu.GetRoom(roomID)
	if !ok {
		return
	}

	room.RemoveRelay(relayID)
}

func (s *Signaling) connectRelay(ctx context.Context, room *sfu.Room, owner string) error {
	_, err, _ := s.relayGroup.Do(room.ID(), func() (any, error) {
		if _, ok := room.GetRelay(owner); ok {
			return nil, nil
		}

		offer, err := room.CreateRelayOffer(owner)
		if err != nil {
			return nil, fmt.Errorf("create relay offer: %w", err)
		}

		resp, err := s.relay.ConnectRelay(ctx, owner, &gnpb.ConnectRelayRequest{
			RoomId:  room.ID(),
			RelayId: s.self,
			Sdp:     offer.SDP,
		})
		if err != nil {
			room.RemoveRelay(owner)
			return nil, fmt.Errorf("relay connect: %w", err)
		}

		err = room.SetRelayAnswer(owner, webrtc.SessionDescription{
			Type: webrtc.SDPTypeAnswer,
			SDP:  resp.Sdp,
		})
		if err != nil {
			room.RemoveRelay(owner)
			return nil, fmt.Errorf("set relay answer: %w", err)
		}

		return nil, nil
	})

	return err
}

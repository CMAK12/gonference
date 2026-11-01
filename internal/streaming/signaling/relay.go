package signaling

import (
	"context"
	"errors"
	"fmt"

	gnpb "github.com/CMAK12/gonference/internal/gen/streaming/v1"
	"github.com/CMAK12/gonference/internal/streaming/entity"
	"github.com/CMAK12/gonference/internal/streaming/sfu"
	"github.com/CMAK12/gonference/internal/streaming/sfu/peers"

	"github.com/pion/webrtc/v4"
)

const relayPrefix = "relay-"

func (s *Signaling) HandleRelayOffer(ctx context.Context, offer entity.RelayOffer) (entity.RelayAnswer, error) {
	room := s.sfu.GetOrCreateRoom(offer.RoomID)

	relay, err := sfu.NewPeer(room, relayPeerID(offer.RelayID), peers.NewRelayPeer)
	if err != nil {
		return entity.RelayAnswer{}, fmt.Errorf("signaling.HandleRelayOffer: failed to create relay peer: %w", err)
	}

	answer, err := relay.CreateAnswer(webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  offer.SDP,
	})
	if err != nil {
		return entity.RelayAnswer{}, fmt.Errorf("signaling.HandleRelayOffer: failed to create answer: %w", errors.Join(err, relay.Close()))
	}

	room.Attach(relay)

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

	room.RemovePeer(relayPeerID(relayID))
}

func (s *Signaling) connectRelay(ctx context.Context, room *sfu.Room, owner string) error {
	_, err, _ := s.relayGroup.Do(room.ID(), func() (any, error) {
		id := relayPeerID(owner)
		if _, ok := room.GetPeer(id); ok {
			return nil, nil
		}

		relay, err := sfu.NewPeer(room, id, peers.NewRelayPeer)
		if err != nil {
			return nil, fmt.Errorf("create relay peer: %w", err)
		}

		room.Attach(relay)

		offer, err := relay.CreateOffer()
		if err != nil {
			room.RemovePeer(id)
			return nil, fmt.Errorf("create relay offer: %w", err)
		}

		resp, err := s.relay.ConnectRelay(ctx, owner, &gnpb.ConnectRelayRequest{
			RoomId:  room.ID(),
			RelayId: s.self,
			Sdp:     offer.SDP,
		})
		if err != nil {
			room.RemovePeer(id)
			return nil, fmt.Errorf("relay connect: %w", err)
		}

		err = relay.SetAnswer(webrtc.SessionDescription{
			Type: webrtc.SDPTypeAnswer,
			SDP:  resp.Sdp,
		})
		if err != nil {
			room.RemovePeer(id)
			return nil, fmt.Errorf("set relay answer: %w", err)
		}

		return nil, nil
	})

	return err
}

func relayPeerID(server string) string {
	return relayPrefix + server
}

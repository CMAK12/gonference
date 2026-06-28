package signaling

import (
	"fmt"

	"github.com/CMAK12/gonference/internal/streaming/entity"
	"github.com/CMAK12/gonference/internal/streaming/sfu"

	"github.com/pion/webrtc/v3"
)

type (
	SFU interface {
		GetRoom(id string) (*sfu.Room, bool)
		GetOrCreateRoom(id string) *sfu.Room
	}

	Transport interface {
		WriteMessage(data entity.SignalMessage) error
		Close() error
	}
)

type Signaling struct {
	sfu SFU
}

func New(sfu SFU) *Signaling {
	return &Signaling{
		sfu: sfu,
	}
}

func (s *Signaling) HandleOffer(message entity.SignalMessage, tn Transport) error {
	room := s.sfu.GetOrCreateRoom(message.RoomID)

	offer := webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  *message.SDP,
	}

	_, err := room.AddPeer(tn, offer, message.MemberID)
	if err != nil {
		return fmt.Errorf("signaling.HanleOffer: failed to add peer: %v", err)
	}

	return nil
}

func (s *Signaling) HandleAnswer(message entity.SignalMessage) error {
	room, ok := s.sfu.GetRoom(message.RoomID)
	if !ok {
		return fmt.Errorf("signaling.HandleAnswer: room not found: %s", message.MemberID)
	}

	peer, ok := room.GetPeer(message.MemberID)
	if !ok {
		return fmt.Errorf("signaling.HandleAnswer: peer not found: %s", message.MemberID)
	}

	return peer.ValidateAnswer(webrtc.SessionDescription{
		Type: webrtc.SDPTypeAnswer,
		SDP:  *message.SDP,
	})
}

func (s *Signaling) HandleCandidate(message entity.SignalMessage) error {
	room, ok := s.sfu.GetRoom(message.RoomID)
	if !ok {
		return fmt.Errorf("signaling.HandleCandidate: room not found: %s", message.MemberID)
	}

	peer, ok := room.GetPeer(message.MemberID)
	if !ok {
		return fmt.Errorf("signaling.HandleCandidate: peer not found: %s", message.MemberID)
	}

	if err := peer.AddICECandidate(*message.Candidate); err != nil {
		return fmt.Errorf("signaling.HandleCandidate: failed to add ICE candidate: %v", err)
	}

	return nil
}

func (s *Signaling) HandleLeave(message entity.SignalMessage) error {
	room, ok := s.sfu.GetRoom(message.RoomID)
	if !ok {
		return fmt.Errorf("signaling.HandleLeave: room not found: %s", message.RoomID)
	}

	peer, ok := room.GetPeer(message.MemberID)
	if !ok {
		return fmt.Errorf("signaling.HandleLeave: peer not found: %s", message.MemberID)
	}

	if err := peer.Close(); err != nil {
		return fmt.Errorf("signaling.HandleLeave: failed to close peer connection: %v", err)
	}
	return nil
}

package signaling

import (
	"fmt"
	"gonference/internal/entity"
	"gonference/internal/usecase/sfu"

	"github.com/pion/webrtc/v3"
)

type (
	SFU interface {
		GetRoom(id string) (*sfu.Room, bool)
		GetOrCreateRoom(id string) *sfu.Room
	}

	Transport interface {
		WriteMessage(data entity.Message) error
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

func (s *Signaling) HandleOffer(message entity.Message, tn Transport) error {
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

func (s *Signaling) HandleAnswer(message entity.Message) error {
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

func (s *Signaling) HandleCandidate(message entity.Message) error {
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

func (s *Signaling) HandleBye(message entity.Message) error {
	room, ok := s.sfu.GetRoom(message.RoomID)
	if !ok {
		return fmt.Errorf("signaling.HandleBye: room not found: %s", message.RoomID)
	}

	peer, ok := room.GetPeer(message.MemberID)
	if !ok {
		return fmt.Errorf("signaling.HandleBye: peer not found: %s", message.MemberID)
	}

	if err := peer.Close(); err != nil {
		return fmt.Errorf("signaling.HandleBye: failed to close peer connection: %v", err)
	}
	return nil
}

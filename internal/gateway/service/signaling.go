package service

import (
	"context"

	signalpb "github.com/CMAK12/gonference/internal/gen/streaming/v1"
)

func (s *Service) Connect(ctx context.Context, req *signalpb.SignalMessage) (*signalpb.SignalMessage, error) {
	resp, err := s.signaling.Connect(ctx, &signalpb.SignalMessage{
		Type:      req.GetType(),
		RoomId:    req.GetRoomId(),
		MemberId:  req.GetMemberId(),
		Sdp:       req.Sdp,
		Candidate: req.GetCandidate(),
	})
	if err != nil {
		return nil, err
	}

	return resp, nil
}

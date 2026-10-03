package service

import (
	"context"

	pb "github.com/CMAK12/gonference/internal/gen/streaming/v1"
	signalpb "github.com/CMAK12/gonference/internal/gen/streaming/v1"
)

func (s *Service) Connect(ctx context.Context, req *pb.SignalMessage) (*pb.SignalMessage, error) {
	resp, err := s.signaling.Connect(ctx, signalpb.SignalMessage{
		Type:      req.GetType(),
		RoomId:    req.GetRoomId(),
		Sdp:       new(req.GetSdp()),
		Candidate: req.GetCandidate(),
	})
	if err != nil {
		return nil, err
	}

	return resp, nil
}

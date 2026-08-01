package service

import (
	"context"
	"fmt"

	"github.com/CMAK12/gonference/internal/gateway/service/mapper"
	pb "github.com/CMAK12/gonference/internal/gen/gateway/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Service) CreateConference(ctx context.Context, req *pb.CreateConferenceRequest) (*pb.CreateConferenceResponse, error) {
	resp, err := s.conference.CreateConference(ctx, mapper.ToConferenceCreateRequest(req))
	if err != nil {
		return nil, fmt.Errorf("create conference: %w", err)
	}

	return mapper.FromConferenceCreateResponse(resp), nil
}

func (s *Service) JoinConference(ctx context.Context, req *pb.JoinConferenceRequest) (*pb.JoinConferenceResponse, error) {
	resp, err := s.conference.JoinConference(ctx, mapper.ToConferenceJoinRequest(req))
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, fmt.Errorf("join conference %s: %w", req.GetConferenceId(), ErrConferenceNotFound)
		}

		return nil, fmt.Errorf("join conference %s: %w", req.GetConferenceId(), err)
	}

	return mapper.FromConferenceJoinResponse(resp), nil
}

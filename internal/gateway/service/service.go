package service

import (
	"context"

	confpb "github.com/CMAK12/gonference/internal/gen/conference/v1"
)

type ConferenceClient interface {
	CreateConference(ctx context.Context, req *confpb.CreateConferenceRequest) (*confpb.CreateConferenceResponse, error)
	JoinConference(ctx context.Context, req *confpb.JoinConferenceRequest) (*confpb.JoinConferenceResponse, error)
}

type Service struct {
	conference ConferenceClient
}

func New(conference ConferenceClient) *Service {
	return &Service{
		conference: conference,
	}
}

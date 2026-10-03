package service

import (
	"context"

	confpb "github.com/CMAK12/gonference/internal/gen/conference/v1"
	signalpb "github.com/CMAK12/gonference/internal/gen/streaming/v1"
)

type ConferenceClient interface {
	CreateConference(ctx context.Context, req *confpb.CreateConferenceRequest) (*confpb.CreateConferenceResponse, error)
	JoinConference(ctx context.Context, req *confpb.JoinConferenceRequest) (*confpb.JoinConferenceResponse, error)
}

type SignalingClient interface {
	Connect(ctx context.Context, req *signalpb.SignalMessage) (*signalpb.SignalMessage, error)
}

type Service struct {
	conference ConferenceClient
	signaling  SignalingClient
}

func New(conference ConferenceClient, signaling SignalingClient) *Service {
	return &Service{
		conference: conference,
		signaling:  signaling,
	}
}

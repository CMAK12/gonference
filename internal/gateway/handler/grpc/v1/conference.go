package grpcv1

import (
	"context"
	"errors"
	"log/slog"

	"github.com/CMAK12/gonference/internal/gateway/service"
	pb "github.com/CMAK12/gonference/internal/gen/gateway/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ConferenceServer struct {
	pb.UnimplementedConferenceServer

	log *slog.Logger

	src Service
}

func registerConferenceServer(server grpc.ServiceRegistrar, service Service) *ConferenceServer {
	conference := &ConferenceServer{
		log: slog.Default().With(slog.String("component", "conference-grpc")),
		src: service,
	}

	pb.RegisterConferenceServer(server, conference)

	return conference
}

func (s *ConferenceServer) CreateConference(ctx context.Context, req *pb.CreateConferenceRequest) (*pb.CreateConferenceResponse, error) {
	resp, err := s.src.CreateConference(ctx, req)
	if err != nil {
		s.log.Error("Failed to create conference", slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, "could not create conference")
	}

	return resp, nil
}

func (s *ConferenceServer) JoinConference(ctx context.Context, req *pb.JoinConferenceRequest) (*pb.JoinConferenceResponse, error) {
	resp, err := s.src.JoinConference(ctx, req)
	if err != nil {
		if errors.Is(err, service.ErrConferenceNotFound) {
			return nil, status.Error(codes.NotFound, "conference not found")
		}

		s.log.Error("Failed to join conference", slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, "could not join conference")
	}

	return resp, nil
}

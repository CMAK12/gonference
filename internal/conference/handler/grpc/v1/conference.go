package grpcv1

import (
	"context"
	"log/slog"

	pb "github.com/CMAK12/gonference/internal/gen/conference/v1"
	"google.golang.org/grpc"
)

var _ pb.ConferenceServer = (*ConferenceServer)(nil)

type ConferenceService interface {
}

type ConferenceServer struct {
	pb.UnimplementedConferenceServer

	log *slog.Logger

	src ConferenceService
}

func registerConferenceServer(server grpc.ServiceRegistrar, cs ConferenceService) *ConferenceServer {
	confServer := &ConferenceServer{
		log: slog.Default().With(slog.String("component", "conference-grpc")),
		src: cs,
	}

	pb.RegisterConferenceServer(server, confServer)

	return confServer
}

func (s *ConferenceServer) CreateConference(ctx context.Context, req *pb.CreateConferenceRequest) (*pb.CreateConferenceResponse, error) {
	return nil, nil
}

func (s *ConferenceServer) JoinConference(ctx context.Context, req *pb.JoinConferenceRequest) (*pb.JoinConferenceResponse, error) {
	return nil, nil
}

package grpcv1

import (
	"context"
	"errors"
	"log/slog"

	"github.com/CMAK12/gonference/internal/conference/entity"
	"github.com/CMAK12/gonference/internal/conference/entity/mapper"
	"github.com/CMAK12/gonference/internal/conference/storage"
	pb "github.com/CMAK12/gonference/internal/gen/conference/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ pb.ConferenceServer = (*ConferenceServer)(nil)

type ConferenceService interface {
	CreateConference(ctx context.Context, conf *entity.Conference) error
	JoinConference(ctx context.Context, conf *entity.Conference) error
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
	conf := mapper.CreateRequestToConference(req)

	if err := s.src.CreateConference(ctx, conf); err != nil {
		s.log.Error("Failed to create conference", slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, "could not create conference")
	}

	return &pb.CreateConferenceResponse{
		ConferenceId: conf.ID,
	}, nil
}

func (s *ConferenceServer) JoinConference(ctx context.Context, req *pb.JoinConferenceRequest) (*pb.JoinConferenceResponse, error) {
	conf := &entity.Conference{ID: req.GetConferenceId()}

	if err := s.src.JoinConference(ctx, conf); err != nil {
		if errors.Is(err, storage.ErrConferenceNotFound) {
			return nil, status.Error(codes.NotFound, "conference not found")
		}
		s.log.Error("Failed to join conference", slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, "could not join conference")
	}

	return &pb.JoinConferenceResponse{
		ConferenceId: conf.ID,
		MemberId:     req.GetMemberId(),
		// Url is populated once the streaming service integration lands.
	}, nil
}

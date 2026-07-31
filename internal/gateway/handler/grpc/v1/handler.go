package grpcv1

import (
	"context"

	pb "github.com/CMAK12/gonference/internal/gen/gateway/v1"
	"google.golang.org/grpc"
)

type Service interface {
	CreateConference(ctx context.Context, req *pb.CreateConferenceRequest) (*pb.CreateConferenceResponse, error)
	JoinConference(ctx context.Context, req *pb.JoinConferenceRequest) (*pb.JoinConferenceResponse, error)
}

func RegisterGRPCV1Handler(server grpc.ServiceRegistrar, cs Service) {
	registerConferenceServer(server, cs)
	registerSignalingServer(server)
}

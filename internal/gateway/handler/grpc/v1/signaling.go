package grpcv1

import (
	"context"
	"log/slog"

	pb "github.com/CMAK12/gonference/internal/gen/gateway/v1"
	"google.golang.org/grpc"
)

type SignalingServer struct {
	pb.UnimplementedSignalingServer

	log *slog.Logger
}

func registerSignalingServer(server grpc.ServiceRegistrar) *SignalingServer {
	signaling := &SignalingServer{
		log: slog.Default().With(slog.String("component", "signaling")),
	}

	pb.RegisterSignalingServer(server, signaling)

	return signaling
}

func (s *SignalingServer) Connect(ctx context.Context, req *pb.SignalMessage) (*pb.SignalMessage, error) {
	s.log.Info("Received signaling message", slog.Any("message", req))
	return nil, nil
}

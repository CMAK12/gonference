package grpcv1

import (
	"context"
	"log/slog"

	gnpb "github.com/CMAK12/gonference/internal/gen/streaming/v1"
	"github.com/CMAK12/gonference/internal/streaming/entity/mapper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ gnpb.SignalingServer = (*SignalingServer)(nil)

type SignalingServer struct {
	gnpb.UnimplementedSignalingServer

	log *slog.Logger

	signaling Signaling
}

func registerSignalingServer(server grpc.ServiceRegistrar, s Signaling) *SignalingServer {
	ss := &SignalingServer{
		log:       slog.Default().With(slog.String("component", "signaling-grpc")),
		signaling: s,
	}

	gnpb.RegisterSignalingServer(server, ss)

	return ss
}

func (ss *SignalingServer) Connect(ctx context.Context, req *gnpb.SignalMessage) (*gnpb.SignalMessage, error) {
	answer, err := ss.signaling.HandleOffer(ctx, mapper.MapProtoToMessage(req))
	if err != nil {
		ss.log.Error("Failed to handle offer", slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, "could not handle offer")
	}

	return mapper.MapMessageToProto(answer), nil
}

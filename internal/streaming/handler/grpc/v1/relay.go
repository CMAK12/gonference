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

var _ gnpb.RelayServer = (*RelayServer)(nil)

type RelayServer struct {
	gnpb.UnimplementedRelayServer

	log *slog.Logger

	relay Relay
}

func registerRelayServer(server grpc.ServiceRegistrar, r Relay) *RelayServer {
	rs := &RelayServer{
		log:   slog.Default().With(slog.String("component", "relay-grpc")),
		relay: r,
	}

	gnpb.RegisterRelayServer(server, rs)

	return rs
}

func (rs *RelayServer) Connect(ctx context.Context, req *gnpb.ConnectRelayRequest) (*gnpb.ConnectRelayResponse, error) {
	answer, err := rs.relay.HandleRelayOffer(ctx, mapper.MapProtoToRelayOffer(req))
	if err != nil {
		rs.log.Error("Failed to handle relay offer",
			slog.String("room_id", req.RoomId),
			slog.String("relay_id", req.RelayId),
			slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, "could not handle relay offer")
	}

	return mapper.MapRelayAnswerToProto(answer), nil
}

func (rs *RelayServer) Disconnect(ctx context.Context, req *gnpb.DisconnectRelayRequest) (*gnpb.DisconnectRelayResponse, error) {
	rs.relay.HandleRelayLeave(ctx, req.RoomId, req.RelayId)

	return &gnpb.DisconnectRelayResponse{}, nil
}

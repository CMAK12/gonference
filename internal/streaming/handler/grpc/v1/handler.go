package grpcv1

import (
	"context"

	"github.com/CMAK12/gonference/internal/streaming/entity"
	"google.golang.org/grpc"
)

type (
	Signaling interface {
		HandleOffer(ctx context.Context, message entity.SignalMessage) (entity.SignalMessage, error)
	}

	Relay interface {
		HandleRelayOffer(ctx context.Context, offer entity.RelayOffer) (entity.RelayAnswer, error)
		HandleRelayLeave(ctx context.Context, roomID, relayID string)
	}
)

func RegisterGRPCV1Handler(server grpc.ServiceRegistrar, s Signaling, r Relay) {
	registerSignalingServer(server, s)
	registerRelayServer(server, r)
}

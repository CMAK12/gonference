package grpcv1

import (
	"context"

	"github.com/CMAK12/gonference/internal/streaming/entity"
	"google.golang.org/grpc"
)

type Signaling interface {
	HandleOffer(ctx context.Context, message entity.SignalMessage) (entity.SignalMessage, error)
}

func RegisterGRPCV1Handler(server grpc.ServiceRegistrar, s Signaling) {
	registerSignalingServer(server, s)
}

package grpcv1

import (
	"context"
	"fmt"
	"log/slog"

	impb "github.com/CMAK12/gonference/internal/gen/streaming/v1"
	"github.com/CMAK12/gonference/internal/streaming/entity"
	"github.com/CMAK12/gonference/internal/streaming/entity/mapper"
	"google.golang.org/grpc"
)

var _ impb.SignalingServer = (*SignalingServer)(nil)

type Signaling interface {
	HandleOffer(message entity.SignalMessage) (entity.SignalMessage, error)
}

type SignalingServer struct {
	impb.UnimplementedSignalingServer

	log *slog.Logger

	signaling Signaling
}

func registerSignalingServer(server grpc.ServiceRegistrar, s Signaling) *SignalingServer {
	ss := &SignalingServer{
		log:       slog.Default().With(slog.String("component", "signaling-grpc")),
		signaling: s,
	}

	impb.RegisterSignalingServer(server, ss)

	return ss
}

// Connect bootstraps a peer: it accepts the client's offer and returns the SFU's
// answer. Subsequent signaling is exchanged over the peer's DataChannel.
func (ss *SignalingServer) Connect(_ context.Context, req *impb.SignalMessage) (*impb.SignalMessage, error) {
	if req.Type != impb.MessageType_OFFER {
		return nil, fmt.Errorf("grpcv1.Connect: expected offer, got %s", req.Type.String())
	}

	answer, err := ss.signaling.HandleOffer(mapper.MapProtoToMessage(req))
	if err != nil {
		ss.log.Error("Failed to handle offer", slog.String("error", err.Error()))
		return nil, fmt.Errorf("grpcv1.Connect: could not handle offer: %w", err)
	}

	return mapper.MapMessageToProto(answer), nil
}

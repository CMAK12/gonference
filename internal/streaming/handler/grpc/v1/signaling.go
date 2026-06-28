package grpcv1

import (
	"errors"
	"fmt"
	"io"
	"log/slog"

	impb "github.com/CMAK12/gonference/internal/gen/streaming/v1"
	"github.com/CMAK12/gonference/internal/streaming/entity"
	"github.com/CMAK12/gonference/internal/streaming/entity/mapper"
	"google.golang.org/grpc"
)

var _ impb.SignalingServer = (*SignalingServer)(nil)

type Signaling interface {
	HandleOffer(message entity.SignalMessage) error
	HandleAnswer(message entity.SignalMessage) error
	HandleCandidate(message entity.SignalMessage) error
	HandleLeave(message entity.SignalMessage) error
}

type SignalingServer struct {
	impb.UnimplementedSignalingServer

	log *slog.Logger

	signaling Signaling
}

func NewSignalingServer(s Signaling) *SignalingServer {
	return &SignalingServer{
		signaling: s,
	}
}

func (ss *SignalingServer) Connect(stream grpc.BidiStreamingServer[impb.SignalMessage, impb.SignalMessage]) error {
	for {
		req, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}

			return fmt.Errorf("grpcv1.Connect: could not receive a request: %w", err)
		}

		msg := mapper.MapProtoToMessage(req)

		switch req.Type {
		case impb.MessageType_OFFER:
			if err := ss.signaling.HandleOffer(msg); err != nil {
				return fmt.Errorf("grpcv1.Connect: could not handle offer: %w", err)
			}

		case impb.MessageType_ANSWER:
			if err := ss.signaling.HandleAnswer(msg); err != nil {
				return fmt.Errorf("grpcv1.Connect: could not handle answer: %w", err)
			}

		case impb.MessageType_CANDIDATE:
			if err := ss.signaling.HandleCandidate(msg); err != nil {
				return fmt.Errorf("grpcv1.Connect: could not handle candidate: %w", err)
			}

		case impb.MessageType_LEAVE:
			if err := ss.signaling.HandleLeave(msg); err != nil {
				return fmt.Errorf("grpcv1.Connect: could not handle leave: %w", err)
			}

		default:
			ss.log.Warn("Unknown message type", slog.String("type", req.Type.String()))
		}
	}
}

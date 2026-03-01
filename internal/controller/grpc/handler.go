package grpc

import (
	"gonference/internal/entity"
	"gonference/internal/usecase/signaling"
	"gonference/pkg/pb"
	"log/slog"
)

type SignalingService interface {
	HandleOffer(message entity.Message, signaling signaling.Transport) error
	HandleAnswer(message entity.Message) error
	HandleCandidate(message entity.Message) error
}

type Handler struct {
	logger *slog.Logger

	pb.UnimplementedSignalingServer
	signaling SignalingService
}

func NewHandler(signaling SignalingService) *Handler {
	return &Handler{
		logger:    slog.Default().With(slog.String("component", "grpc")),
		signaling: signaling,
	}
}

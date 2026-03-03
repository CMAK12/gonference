package websocket

import (
	"gonference/internal/entity"
	"gonference/internal/usecase/sfu"
	"gonference/internal/usecase/signaling"
	"log/slog"
)

type (
	SFU interface {
		GetOrCreateRoom(id string) *sfu.Room
	}

	SignalingService interface {
		HandleOffer(message entity.Message, transport signaling.Transport) error
		HandleAnswer(message entity.Message) error
		HandleCandidate(message entity.Message) error
		HandleLeave(message entity.Message) error
	}
)

type Handler struct {
	logger    *slog.Logger
	sfu       SFU
	signaling SignalingService
}

func NewHandler(sfu SFU, signaling SignalingService) *Handler {
	return &Handler{
		logger:    slog.Default().With(slog.String("component", "ws")),
		sfu:       sfu,
		signaling: signaling,
	}
}

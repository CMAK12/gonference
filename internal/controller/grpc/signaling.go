package grpc

import (
	"errors"
	"gonference/internal/entity"
	"gonference/pkg/pb"
	"io"
	"log/slog"
)

func (h *Handler) Connect(stream pb.Signaling_ConnectServer) error {
	transport := NewTransport(stream)

	for {
		message, err := transport.ReadMessage()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			h.logger.Error("Failed to read message", slog.String("error", err.Error()))
			return err
		}

		switch message.Type {
		case entity.TypeOffer:
			if err := h.signaling.HandleOffer(message, transport); err != nil {
				h.logger.Error("Failed to handle offer", slog.String("error", err.Error()))
			}
		case entity.TypeAnswer:
			if err := h.signaling.HandleAnswer(message); err != nil {
				h.logger.Error("Failed to handle answer", slog.String("error", err.Error()))
			}
		case entity.TypeCandidate:
			if err := h.signaling.HandleCandidate(message); err != nil {
				h.logger.Error("Failed to handle candidate", slog.String("error", err.Error()))
			}
		case entity.TypeLeave:
			//if err := h.signaling.HandleLeave(message); err != nil {
			//}
		default:
			h.logger.Error("Unknown message type", slog.String("type", string(message.Type)))
			continue
		}
	}
}

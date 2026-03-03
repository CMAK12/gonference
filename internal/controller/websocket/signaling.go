package websocket

import (
	"errors"
	"gonference/internal/entity"
	"log/slog"
	"net/http"
)

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	ws, err := NewWebSocket(w, r)
	if err != nil {
		h.logger.Error("Failed to upgrade to WebSocket", slog.String("error", err.Error()))
		return
	}
	defer func() {
		if err := ws.Close(); err != nil {
			h.logger.Error("Failed to close WebSocket", slog.String("error", err.Error()))
		}
	}()

	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			if errors.Is(err, ErrConnectionClosed) {
				break
			}
			h.logger.Error("Failed to read message", slog.String("error", err.Error()))
			break
		}

		switch msg.Type {
		case entity.TypeOffer:
			if err := h.signaling.HandleOffer(msg, ws); err != nil {
				h.logger.Error("Failed to handle offer", slog.String("error", err.Error()))
			}
		case entity.TypeAnswer:
			if err := h.signaling.HandleAnswer(msg); err != nil {
				h.logger.Error("Failed to handle answer", slog.String("error", err.Error()))
			}
		case entity.TypeCandidate:
			if err := h.signaling.HandleCandidate(msg); err != nil {
				h.logger.Error("Failed to handle candidate", slog.String("error", err.Error()))
			}
		case entity.TypeLeave:
			if err := h.signaling.HandleLeave(msg); err != nil {
				h.logger.Error("Failed to handle leave", slog.String("error", err.Error()))
			}
		default:
			h.logger.Error("Unknown message type", slog.String("type", string(msg.Type)))
			continue
		}
	}
}

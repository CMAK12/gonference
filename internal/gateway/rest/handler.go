package rest

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/CMAK12/gonference/internal/gateway/config"
)

// Signaling forwards a publisher's SDP offer to the streaming service and
// returns the answer SDP. It is the only dependency the WHIP endpoint needs.
type Signaling interface {
	Connect(ctx context.Context, roomID, memberID, offer string) (string, error)
}

type Handler struct {
	logger    *slog.Logger
	signaling Signaling
	srv       *http.Server
}

func NewHandler(cfg config.REST, signaling Signaling) *Handler {
	logger := slog.Default().With(slog.String("component", "rest"))

	mux := http.NewServeMux()
	handler := withLogging(mux, logger)
	handler = withCORS(handler)

	h := &Handler{
		logger:    logger,
		signaling: signaling,
		srv: &http.Server{
			Addr:    fmt.Sprintf(":%d", cfg.Port),
			Handler: handler,
		},
	}

	mux.HandleFunc("POST /whip", h.handleWHIP)

	return h
}

func (h *Handler) ListenAndServe() {
	h.logger.Info("started", slog.String("addr", h.srv.Addr))

	if err := h.srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		h.logger.Error("during serving", slog.String("error", err.Error()))
	}
}

func (h *Handler) Close() {
	if err := h.srv.Close(); err != nil {
		h.logger.Error("during closing", slog.String("error", err.Error()))
	}

	h.logger.Info("stopped")
}

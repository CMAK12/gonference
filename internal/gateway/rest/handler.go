package rest

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/CMAK12/gonference/internal/gateway/config"
)

type Handler struct {
	logger *slog.Logger
	srv    *http.Server
}

func NewHandler(cfg config.REST) *Handler {
	logger := slog.Default().With(slog.String("component", "rest"))

	mux := http.NewServeMux()
	handler := withLogging(mux, logger)
	handler = withCORS(handler)

	h := &Handler{
		logger: logger,
		srv: &http.Server{
			Addr:    fmt.Sprintf(":%d", cfg.Port),
			Handler: handler,
		},
	}

	mux.HandleFunc("GET /whep", h.getWHEP)
	mux.HandleFunc("POST /whep", h.handleWHEP)

	mux.HandleFunc("POST /conference/create", h.createConference)
	mux.HandleFunc("GET /conference", h.listConferences)
	mux.HandleFunc("GET /conference/{id}/join", h.joinConference)
	mux.HandleFunc("DELETE /conference/{id}/leave", h.removeMember)

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

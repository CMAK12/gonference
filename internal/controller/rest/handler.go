package rest

import (
	"errors"
	"fmt"
	"gonference/internal/sfu"
	"log/slog"
	"net/http"

	"gonference/internal/config"
	"gonference/internal/controller/middleware"
)

type SFU interface {
	GetOrCreateRoom(id string) *sfu.Room
	Close()
}

type Handler struct {
	logger *slog.Logger
	srv    *http.Server

	sfu SFU
}

func NewHandler(cfg config.REST, sfu SFU) *Handler {
	logger := slog.Default().With(slog.String("component", "rest"))

	mux := http.NewServeMux()
	handler := middleware.WithLogging(mux, logger)
	handler = middleware.WithCORS(handler)

	h := &Handler{
		logger: logger,
		srv: &http.Server{
			Addr:    fmt.Sprintf(":%d", cfg.Port),
			Handler: handler,
		},
		sfu: sfu,
	}

	mux.HandleFunc("GET /whep", h.getWHEP)
	mux.HandleFunc("POST /whep", h.handleWHEP)

	mux.HandleFunc("GET /ws", h.wsHandler)

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
	h.sfu.Close()
	if err := h.srv.Close(); err != nil {
		h.logger.Error("during closing", slog.String("error", err.Error()))
	}

	h.logger.Info("stopped")
}

package rest

import (
	"errors"
	"fmt"
	"gonference/internal/usecase"
	"log/slog"
	"net/http"

	"gonference/internal/config"
	ws "gonference/internal/controller/websocket"
)

type Handler struct {
	logger *slog.Logger
	srv    *http.Server

	uc *usecase.UseCase
}

func NewHandler(cfg config.REST, uc *usecase.UseCase) *Handler {
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
		uc: uc,
	}

	mux.HandleFunc("GET /whep", h.getWHEP)
	mux.HandleFunc("POST /whep", h.handleWHEP)

	wsHandler := ws.NewHandler(uc.SFU, uc.Signaling)
	mux.HandleFunc("GET /ws", wsHandler.Handle)

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
	h.uc.SFU.Close()
	if err := h.srv.Close(); err != nil {
		h.logger.Error("during closing", slog.String("error", err.Error()))
	}

	h.logger.Info("stopped")
}

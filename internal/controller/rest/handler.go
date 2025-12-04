package rest

import (
	"errors"
	"fmt"
	"gonference/internal/config"
	"gonference/internal/controller/middleware"
	"log/slog"
	"net/http"
)

type Handler struct {
	logger *slog.Logger
	srv    *http.Server
}

func NewHandler(cfg config.REST) *Handler {
	logger := slog.Default().With(slog.String("component", "rest"))

	mux := http.NewServeMux()
	handler := middleware.WithLogging(mux, logger)
	handler = middleware.WithCORS(handler)

	h := &Handler{logger: logger, srv: &http.Server{Addr: fmt.Sprintf(":%d", cfg.Port), Handler: handler}}

	mux.HandleFunc("GET /whep", h.getWHEP)
	mux.HandleFunc("POST /whep", h.handleWHEP)

	return h
}

func (h *Handler) ListenAndServe() {
	h.logger.Info("started", slog.String("addr", h.srv.Addr))

	if err := h.srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		h.logger.Error(err.Error())
	}
}

func (h *Handler) Close() {
	if err := h.srv.Close(); err != nil {
		h.logger.Error(err.Error())
	}

	h.logger.Info("stopped")
}

func (h *Handler) handleWHEP(w http.ResponseWriter, r *http.Request) {
	return
}

func (h *Handler) getWHEP(w http.ResponseWriter, r *http.Request) {
	return
}

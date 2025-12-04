package admin_panel

import (
	"embed"
	"errors"
	"fmt"
	"gonference/internal/config"
	"html/template"
	"log/slog"
	"net/http"
)

//go:embed templates
var templatesFs embed.FS

var templates = template.Must(template.New("").ParseFS(templatesFs, "**/*.html"))

type AdminPanelHandler struct {
	logger *slog.Logger
	srv    *http.Server
}

func NewHandler(cfg config.AdminPanel) *AdminPanelHandler {
	logger := slog.Default().With(slog.String("component", "admin-panel"))

	mux := http.NewServeMux()

	handler := &AdminPanelHandler{
		logger: logger,
		srv:    &http.Server{Addr: fmt.Sprintf(":%d", cfg.Port), Handler: mux}}

	mux.HandleFunc("/", handler.getIndex)
	mux.HandleFunc("/conference", handler.getConference)

	return handler
}

func (h *AdminPanelHandler) ListenAndServe() {
	h.logger.Info("started", slog.String("addr", h.srv.Addr))

	if err := h.srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		h.logger.Error(err.Error())
	}
}

func (h *AdminPanelHandler) Close() {
	if err := h.srv.Close(); err != nil {
		h.logger.Error(err.Error())
	}

	h.logger.Info("stopped")
}

func (h *AdminPanelHandler) getIndex(w http.ResponseWriter, r *http.Request) {
	_ = templates.ExecuteTemplate(w, "index.html", nil)
}

func (h *AdminPanelHandler) getConference(w http.ResponseWriter, r *http.Request) {
	_ = templates.ExecuteTemplate(w, "webrtc.html", nil)
}

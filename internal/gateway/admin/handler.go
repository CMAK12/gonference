package admin

import (
	"embed"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"

	"github.com/CMAK12/gonference/internal/gateway/config"
)

//go:embed static/*
var statisFS embed.FS

//go:embed templates
var templatesFs embed.FS

var templates = template.Must(template.New("").ParseFS(templatesFs, "**/*.html"))

type Handler struct {
	logger *slog.Logger
	srv    *http.Server
}

func NewHandler(cfg config.Admin) *Handler {
	logger := slog.Default().With(slog.String("component", "admin-panel"))

	mux := http.NewServeMux()

	handler := &Handler{
		logger: logger,
		srv:    &http.Server{Addr: fmt.Sprintf(":%d", cfg.Port), Handler: mux}}

	mux.Handle("/static/", http.FileServer(http.FS(statisFS)))
	mux.HandleFunc("/", handler.getIndex)
	mux.HandleFunc("/conference/{id}", handler.getConference)

	return handler
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

func (h *Handler) getIndex(w http.ResponseWriter, r *http.Request) {
	_ = templates.ExecuteTemplate(w, "index.html", nil)
}

func (h *Handler) getConference(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL.Query().Get("username"))
	_ = templates.ExecuteTemplate(w, "webrtc.html", nil)
}

package admin_panel

import (
	"embed"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"

	"gonference/internal/config"
	"gonference/internal/usecase"
	"gonference/internal/usecase/conference"
)

//go:embed static/*
var statisFS embed.FS

//go:embed templates
var templatesFs embed.FS

var templates = template.Must(template.New("").ParseFS(templatesFs, "**/*.html"))

type AdminPanelHandler struct {
	logger   *slog.Logger
	srv      *http.Server
	uc       *usecase.UseCase
	restPort int
}

type conferenceView struct {
	ID      string
	Name    string
	Members []string
}

type indexData struct {
	Conferences []conferenceView
	RestPort    int
}

type conferencePageData struct {
	RoomID   string
	Username string
	RestPort int
}

func NewHandler(cfg config.Config, uc *usecase.UseCase) *AdminPanelHandler {
	logger := slog.Default().With(slog.String("component", "admin-panel"))

	mux := http.NewServeMux()

	handler := &AdminPanelHandler{
		logger:   logger,
		uc:       uc,
		restPort: cfg.REST.Port,
		srv:      &http.Server{Addr: fmt.Sprintf(":%d", cfg.AdminPanel.Port), Handler: mux},
	}

	mux.Handle("/static/", http.FileServer(http.FS(statisFS)))
	mux.HandleFunc("/", handler.getIndex)
	mux.HandleFunc("/conference/{id}", handler.getConference)

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
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	list := h.uc.Conference.List()
	views := make([]conferenceView, 0, len(list))
	for _, c := range list {
		views = append(views, toView(c))
	}

	data := indexData{Conferences: views, RestPort: h.restPort}
	if err := templates.ExecuteTemplate(w, "index.html", data); err != nil {
		h.logger.Error("rendering index", slog.String("error", err.Error()))
	}
}

func (h *AdminPanelHandler) getConference(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	username := r.URL.Query().Get("username")

	if username == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	data := conferencePageData{RoomID: id, Username: username, RestPort: h.restPort}
	if err := templates.ExecuteTemplate(w, "conference.html", data); err != nil {
		h.logger.Error("rendering conference", slog.String("error", err.Error()))
	}
}

func toView(c conference.Conference) conferenceView {
	members := c.Members
	if members == nil {
		members = []string{}
	}
	return conferenceView{ID: c.ID, Name: c.Name, Members: members}
}

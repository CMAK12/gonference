package rest

import (
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

func (h *Handler) createConference(w http.ResponseWriter, r *http.Request) {
	roomID := uuid.NewString()

	h.hub.AddPeer(roomID, "", nil)

	w.WriteHeader(http.StatusCreated)
	if _, err := w.Write([]byte(roomID)); err != nil {
		h.logger.Error("writing response", slog.String("error", err.Error()))
	}
}

func (h *Handler) joinConference(w http.ResponseWriter, r *http.Request) {
	roomID := r.URL.Query().Get("id")

	h.hub.AddPeer(roomID, "", nil)

	w.WriteHeader(http.StatusOK)
}

// Under question
func (h *Handler) listConferences(w http.ResponseWriter, r *http.Request) {
}

func (h *Handler) removeMember(w http.ResponseWriter, r *http.Request) {
	roomID := r.URL.Query().Get("id")

	h.hub.RemovePeer(roomID, "")

	w.WriteHeader(http.StatusOK)
}

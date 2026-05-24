package rest

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"gonference/internal/usecase/conference"
)

type createConferenceRequest struct {
	Name string `json:"name"`
}

type conferenceResponse struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Members []string `json:"members"`
}

func toResponse(c conference.Conference) conferenceResponse {
	members := c.Members
	if members == nil {
		members = []string{}
	}
	return conferenceResponse{ID: c.ID, Name: c.Name, Members: members}
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(v); err != nil {
		h.logger.Error("encoding json response", slog.String("error", err.Error()))
	}
}

func (h *Handler) writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, conference.ErrNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, conference.ErrUsernameTaken):
		http.Error(w, err.Error(), http.StatusConflict)
	case errors.Is(err, conference.ErrEmptyName), errors.Is(err, conference.ErrEmptyUsername):
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) createConference(w http.ResponseWriter, r *http.Request) {
	var req createConferenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	c, err := h.uc.Conference.Create(req.Name)
	if err != nil {
		h.writeError(w, err)
		return
	}

	h.writeJSON(w, http.StatusCreated, toResponse(c))
}

func (h *Handler) listConferences(w http.ResponseWriter, r *http.Request) {
	list := h.uc.Conference.List()
	out := make([]conferenceResponse, 0, len(list))
	for _, c := range list {
		out = append(out, toResponse(c))
	}
	h.writeJSON(w, http.StatusOK, out)
}

func (h *Handler) joinConference(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	username := r.URL.Query().Get("username")

	c, err := h.uc.Conference.Join(id, username)
	if err != nil {
		h.writeError(w, err)
		return
	}

	h.writeJSON(w, http.StatusOK, toResponse(c))
}

func (h *Handler) removeMember(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	username := r.URL.Query().Get("username")

	empty, err := h.uc.Conference.Leave(id, username)
	if err != nil {
		h.writeError(w, err)
		return
	}

	if empty {
		h.uc.Conference.Delete(id)
		h.uc.SFU.RemoveRoom(id)
	}

	w.WriteHeader(http.StatusNoContent)
}

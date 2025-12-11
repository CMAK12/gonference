package rest

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"gonference/pkg/conference"

	"github.com/pion/webrtc/v3"
)

type Offer struct {
	RoomID string `json:"roomId"`
	SDP    string `json:"sdp"`
}

func (h *Handler) handleWHEP(w http.ResponseWriter, r *http.Request) {
	var offer Offer
	if err := json.NewDecoder(r.Body).Decode(&offer); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	peer, err := conference.NewMember("s")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		h.logger.Error("Failed to create peer", slog.String("error", err.Error()))
		return
	}

	h.hub.AddPeer(offer.RoomID, "", peer)

	answer, err := peer.CreateAnswer(webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  offer.SDP,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		h.logger.Error("Failed to create answer", slog.String("error", err.Error()))
		return
	}

	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write([]byte(answer.SDP))
}

func (h *Handler) getWHEP(w http.ResponseWriter, r *http.Request) {
	_ = w
	_ = r
}

package rest

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

const (
	// maxOfferSize caps the SDP offer body the gateway is willing to read.
	maxOfferSize = 256 << 10 // 256 KiB

	// connectTimeout bounds how long the gateway waits on the streaming service.
	connectTimeout = 15 * time.Second
)

// handleWHIP implements WebRTC-HTTP Ingestion Protocol bootstrap signaling.
//
// The publisher POSTs its SDP offer as `application/sdp`, identifying the room
// and member via query parameters (`?room=<id>&member=<id>`). The gateway
// forwards the offer to the streaming SFU and returns its answer SDP. All
// further signaling (renegotiation as peers join, leave notifications) flows
// over the peer's WebRTC DataChannel.
func (h *Handler) handleWHIP(w http.ResponseWriter, r *http.Request) {
	roomID := r.URL.Query().Get("room")
	memberID := r.URL.Query().Get("member")
	if roomID == "" || memberID == "" {
		http.Error(w, "missing room or member query parameter", http.StatusBadRequest)
		return
	}

	if ct := r.Header.Get("Content-Type"); ct != "" && ct != "application/sdp" {
		http.Error(w, "expected Content-Type: application/sdp", http.StatusUnsupportedMediaType)
		return
	}

	offer, err := io.ReadAll(io.LimitReader(r.Body, maxOfferSize))
	if err != nil {
		http.Error(w, "failed to read offer", http.StatusBadRequest)
		return
	}
	if len(offer) == 0 {
		http.Error(w, "empty offer", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), connectTimeout)
	defer cancel()

	answer, err := h.signaling.Connect(ctx, roomID, memberID, string(offer))
	if err != nil {
		h.logger.Error("WHIP connect failed",
			slog.String("room", roomID),
			slog.String("member", memberID),
			slog.String("error", err.Error()),
		)
		http.Error(w, "failed to establish session", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/sdp")
	w.Header().Set("Location", fmt.Sprintf("/whip?room=%s&member=%s",
		url.QueryEscape(roomID), url.QueryEscape(memberID)))
	w.WriteHeader(http.StatusCreated)

	if _, err := w.Write([]byte(answer)); err != nil {
		h.logger.Error("WHIP write answer failed", slog.String("error", err.Error()))
	}
}

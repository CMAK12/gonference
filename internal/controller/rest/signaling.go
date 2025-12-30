package rest

import (
	"encoding/json"
	"log"
	"log/slog"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // для прикладу
	},
}

type Message struct {
	Type      string                   `json:"type"`
	RoomID    string                   `json:"roomId"`
	MemberID  string                   `json:"memberId"`
	SDP       string                   `json:"sdp,omitempty"`
	Candidate *webrtc.ICECandidateInit `json:"candidate,omitempty"`
}

func (h *Handler) wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade error:", err)
		return
	}
	defer conn.Close()

	for {
		var message Message

		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("read error:", err)
			break
		}

		if err := json.Unmarshal(msg, &message); err != nil {
			log.Println("unmarshal error:", err)
		}

		switch message.Type {
		case "offer":
			room := h.sfu.GetOrCreateRoom(message.RoomID)

			offer := webrtc.SessionDescription{
				Type: webrtc.SDPTypeOffer,
				SDP:  message.SDP,
			}

			_, err := room.AddPeer(conn, offer, message.MemberID)
			if err != nil {
				h.logger.Error("Failed to add peer", slog.String("error", err.Error()))
				return
			}
		case "answer":
			peer, ok := h.sfu.GetOrCreateRoom(message.RoomID).GetPeer(message.MemberID)
			if !ok {
				h.logger.Error("Peer not found", slog.String("memberId", message.MemberID))
				return
			}

			err := peer.ValidateAnswer(webrtc.SessionDescription{
				Type: webrtc.SDPTypeAnswer,
				SDP:  message.SDP,
			})
			if err != nil {
				h.logger.Error("SetRemote(answer) failed", slog.String("error", err.Error()))
			}
		case "candidate":
			peer, ok := h.sfu.GetOrCreateRoom(message.RoomID).GetPeer(message.MemberID)
			if !ok {
				h.logger.Error("Peer not found", slog.String("memberId", message.MemberID))
				return
			}

			if err := peer.AddICECandidate(*message.Candidate); err != nil {
				h.logger.Error("AddICECandidate failed", slog.String("error", err.Error()))
			}
		default:
			h.logger.Info("unknown message type:", message.Type)
		}
	}
}

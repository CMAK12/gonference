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
	Type     string `json:"type"`
	RoomID   string `json:"roomId"`
	MemberID string `json:"memberId"`
	SDP      string `json:"sdp"`
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
			peer, err := h.sfu.AddPeer(conn, message.RoomID, message.MemberID)
			if err != nil {
				h.logger.Error("Failed to add peer", slog.String("error", err.Error()))
				return
			}

			answer, err := peer.CreateAnswer(webrtc.SessionDescription{
				Type: webrtc.SDPTypeOffer,
				SDP:  message.SDP,
			})
			if err != nil {
				h.logger.Error("Failed to create answer", slog.String("error", err.Error()))
				return
			}

			response, err := json.Marshal(Message{
				Type:     "answer",
				RoomID:   message.RoomID,
				MemberID: message.MemberID,
				SDP:      answer.SDP,
			})

			if err := conn.WriteMessage(websocket.TextMessage, response); err != nil {
				log.Println("write error:", err)
				return
			}
		case "answer":
			peer := h.sfu.GetPeer(message.RoomID, message.MemberID)
			if peer == nil {
				return
			}

			err := peer.SetRemoteDescription(webrtc.SessionDescription{
				Type: webrtc.SDPTypeAnswer,
				SDP:  message.SDP,
			})
			if err != nil {
				h.logger.Error("SetRemote(answer) failed", slog.String("error", err.Error()))
			}
		default:
			h.logger.Info("unknown message type:", message.Type)
		}
	}
}

package websocket

import (
	"encoding/json"
	"errors"
	"fmt"
	"gonference/internal/entity"
	"net/http"

	"github.com/gorilla/websocket"
)

var (
	ErrConnectionClosed = errors.New("ws: connection closed")
	ErrDuringClose      = errors.New("ws: error during closing connection")
)

type WebSocket struct {
	conn     *websocket.Conn
	upgrader websocket.Upgrader
}

func NewWebSocket(w http.ResponseWriter, r *http.Request) (*WebSocket, error) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, fmt.Errorf("ws: error during upgrading: %v", err)
	}

	return &WebSocket{
		upgrader: upgrader,
		conn:     conn,
	}, nil
}

func (ws *WebSocket) Close() error {
	if err := ws.conn.Close(); err != nil {
		return ErrDuringClose
	}

	return nil
}

func (ws *WebSocket) ReadMessage() (messageType int, p entity.Message, err error) {
	messageType, body, err := ws.conn.ReadMessage()
	if err != nil {
		if websocket.IsCloseError(
			err,
			websocket.CloseNormalClosure,
			websocket.CloseGoingAway,
		) {
			return 0, entity.Message{}, ErrConnectionClosed
		}
		return 0, entity.Message{}, fmt.Errorf("ws: read error: %v", err)
	}

	var msg entity.Message
	if err := json.Unmarshal(body, &msg); err != nil {
		return 0, entity.Message{}, fmt.Errorf("ws: error during unmarshalling message: %v", err)
	}

	return messageType, msg, nil
}

func (ws *WebSocket) WriteMessage(data entity.Message) error {
	body, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("ws: error during marshaling message: %v", err)
	}

	if err := ws.conn.WriteMessage(websocket.TextMessage, body); err != nil {
		return fmt.Errorf("ws: error during writing message: %v", err)
	}
	return nil
}

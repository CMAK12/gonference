package peers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/CMAK12/gonference/internal/streaming/entity"

	"github.com/pion/webrtc/v4"
)

type Peer struct {
	*connection
}

func NewPeer(api *webrtc.API, roomID, id string, callbacks Callbacks) (*Peer, error) {
	conn, err := newConnection(api, roomID, id, callbacks)
	if err != nil {
		return nil, fmt.Errorf("new connection: %w", err)
	}

	peer := &Peer{
		connection: conn,
	}

	peer.pc.OnDataChannel(func(d *webrtc.DataChannel) {
		if d.Label() != signalingChannelLabel {
			return
		}

		peer.bindSignalingChannel(d)
	})

	_, err = peer.pc.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo, webrtc.RTPTransceiverInit{
		Direction: webrtc.RTPTransceiverDirectionRecvonly,
	})
	if err != nil {
		return nil, errors.Join(fmt.Errorf("add video transceiver: %w", err), conn.close())
	}

	return peer, nil
}

func (p *Peer) NotifyLeave(memberID string) error {
	return p.SendSignal(entity.SignalMessage{
		Type:     entity.TypeLeave,
		RoomID:   p.roomID,
		MemberID: memberID,
	})
}

func (p *Peer) bindSignalingChannel(d *webrtc.DataChannel) {
	p.mux.Lock()
	p.dc = d
	p.mux.Unlock()

	d.OnOpen(p.onSignalingOpen)

	d.OnMessage(func(raw webrtc.DataChannelMessage) {
		var msg entity.SignalMessage
		if err := json.Unmarshal(raw.Data, &msg); err != nil {
			p.logger.Error("Failed to unmarshal signaling message", slog.String("error", err.Error()))
			return
		}

		p.handleSignal(msg)
	})
}

func (p *Peer) handleSignal(msg entity.SignalMessage) {
	switch msg.Type {
	case entity.TypeOffer:
		p.handleOffer(msg)

	case entity.TypeAnswer:
		p.handleAnswer(msg)

	case entity.TypeLeave:
		p.notifyClose()

	default:
		p.logger.Debug("Ignoring signaling message", slog.String("type", string(msg.Type)))
	}
}

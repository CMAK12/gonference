package peers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/CMAK12/gonference/internal/streaming/entity"

	"github.com/pion/webrtc/v4"
)

const relaySignalingChannelID uint16 = 334

type RelayPeer struct {
	*connection
}

func NewRelayOfferer(api *webrtc.API, roomID, id string, callbacks Callbacks) (*RelayPeer, error) {
	return newRelayPeer(api, roomID, id, false, callbacks)
}

func NewRelayAnswerer(api *webrtc.API, roomID, id string, callbacks Callbacks) (*RelayPeer, error) {
	return newRelayPeer(api, roomID, id, true, callbacks)
}

func newRelayPeer(api *webrtc.API, roomID, id string, polite bool, callbacks Callbacks) (*RelayPeer, error) {
	conn, err := newConnection(api, roomID, id, polite, callbacks)
	if err != nil {
		return nil, err
	}

	relay := &RelayPeer{
		connection: conn,
	}

	if err := relay.bindSignalingChannel(); err != nil {
		return nil, errors.Join(fmt.Errorf("bind signaling channel: %w", err), conn.close())
	}

	return relay, nil
}

func (p *RelayPeer) NotifyLeave(memberID string) error {
	return p.SendSignal(entity.SignalMessage{
		Type:     entity.TypeLeave,
		RoomID:   p.roomID,
		MemberID: memberID,
	})
}

func (p *RelayPeer) bindSignalingChannel() error {
	dc, err := p.pc.CreateDataChannel(signalingChannelLabel, &webrtc.DataChannelInit{
		Negotiated: new(true),
		ID:         new(relaySignalingChannelID),
	})
	if err != nil {
		return fmt.Errorf("failed to create signaling data channel: %w", err)
	}

	p.mux.Lock()
	p.dc = dc
	p.mux.Unlock()

	dc.OnOpen(p.onSignalingOpen)
	dc.OnMessage(func(msg webrtc.DataChannelMessage) {
		var signal entity.SignalMessage
		if err := json.Unmarshal(msg.Data, &signal); err != nil {
			p.logger.Error("Failed to unmarshal signal message", slog.String("error", err.Error()))
			return
		}

		p.handleSignal(signal)
	})

	return nil
}

func (p *RelayPeer) handleSignal(message entity.SignalMessage) {
	switch message.Type {
	case entity.TypeOffer:
		p.handleOffer(message)

	case entity.TypeAnswer:
		p.handleAnswer(message)

	case entity.TypeLeave:
		p.logger.Debug("Received leave signal", "member_id", message.MemberID)
		p.callbacks.OnMemberLeave(message.MemberID)
	}
}

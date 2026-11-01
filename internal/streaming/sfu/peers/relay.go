package peers

import (
	"errors"
	"fmt"
	"sync"

	"github.com/CMAK12/gonference/internal/streaming/entity"
	"github.com/pion/webrtc/v4"
)

type RelayPeer struct {
	*connection

	negMux             sync.Mutex
	pendingNegotiation bool
}

func NewRelayPeer(api *webrtc.API, roomID, id string, callbacks Callbacks) (*RelayPeer, error) {
	conn, err := newConnection(api, roomID, id, callbacks)
	if err != nil {
		return nil, err
	}

	relay := &RelayPeer{
		connection: conn,
	}

	relay.bindSignalingChannel()

	return relay, nil
}

func (p *RelayPeer) AddTrackAndRenegotiate(track *webrtc.TrackLocalStaticRTP) error {
	//TODO implement me
	panic("implement me")
}

func (p *RelayPeer) RemoveOutboundTracks(trackIDs []string) error {
	//TODO implement me
	panic("implement me")
}

func (p *RelayPeer) NotifyLeave(memberID string) error {
	return p.SendSignal(entity.SignalMessage{
		Type:     entity.TypeLeave,
		RoomID:   p.roomID,
		MemberID: memberID,
	})
}

func (p *RelayPeer) bindSignalingChannel() {
	//TODO implement me
	panic("implement me")
}

func (p *RelayPeer) handleSignal(message entity.SignalMessage) {
	switch message.Type {
	case entity.TypeLeave:
		p.logger.Debug("Received leave signal", "member_id", message.MemberID)
		p.callbacks.OnMemberLeave(message.MemberID)
	}
}

func (p *RelayPeer) Close() error {
	p.mux.Lock()
	dc := p.dc
	p.dc = nil
	p.mux.Unlock()

	p.negMux.Lock()
	p.pendingNegotiation = false
	p.negMux.Unlock()

	var errs []error
	if dc != nil {
		if err := dc.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close signaling channel: %w", err))
		}
	}

	if err := p.close(); err != nil {
		errs = append(errs, fmt.Errorf("close signaling channel: %w", err))
	}

	return errors.Join(errs...)
}

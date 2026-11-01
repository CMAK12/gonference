package peers

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/pion/webrtc/v4"
)

type RelayPeer struct {
	*connection

	negMux             sync.Mutex
	pendingNegotiation bool
	dc                 *webrtc.DataChannel
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
	//TODO implement me
	panic("implement me")
}

func (p *RelayPeer) bindSignalingChannel() {
	var err error
	p.dc, err = p.pc.CreateDataChannel(signalingChannelLabel, nil)
	if err != nil {
		p.logger.Error("Failed to create data channel", slog.String("error", err.Error()))
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

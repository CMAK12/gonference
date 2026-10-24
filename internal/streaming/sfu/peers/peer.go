package peers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"sync"

	"github.com/CMAK12/gonference/internal/streaming/entity"

	"github.com/pion/webrtc/v4"
)

const signalingChannelLabel = "signaling"

type Peer struct {
	*connection

	negMux             sync.Mutex
	pendingNegotiation bool
	dc                 *webrtc.DataChannel
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

func (p *Peer) AddTrackAndRenegotiate(track *webrtc.TrackLocalStaticRTP) error {
	added, err := p.AddOutboundTrack(track)
	if err != nil {
		return err
	}

	if !added {
		p.logger.Warn("Track already exists", slog.String("track", track.ID()))
		return nil
	}

	return p.Renegotiate()
}

func (p *Peer) Renegotiate() error {
	p.mux.RLock()
	dc := p.dc
	p.mux.RUnlock()

	if dc == nil || dc.ReadyState() != webrtc.DataChannelStateOpen {
		return nil
	}

	p.negMux.Lock()
	defer p.negMux.Unlock()

	if p.pc.SignalingState() != webrtc.SignalingStateStable {
		p.pendingNegotiation = true
		return nil
	}
	p.pendingNegotiation = false

	offer, err := p.CreateOffer()
	if err != nil {
		return err
	}

	msg := entity.SignalMessage{
		Type:     entity.TypeOffer,
		RoomID:   p.roomID,
		MemberID: p.id,
		SDP:      &offer.SDP,
	}

	return p.SendSignal(msg)
}

func (p *Peer) Close() error {
	p.mux.Lock()
	dc := p.dc
	p.dc = nil
	p.mux.Unlock()

	p.negMux.Lock()
	p.pendingNegotiation = false
	p.negMux.Unlock()

	var errs []error
	if dc != nil {
		if err := dc.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
			errs = append(errs, fmt.Errorf("close signaling channel: %w", err))
		}
	}

	if err := p.close(); err != nil && !errors.Is(err, net.ErrClosed) {
		errs = append(errs, fmt.Errorf("close peer connection: %w", err))
	}

	return errors.Join(errs...)
}

func (p *Peer) bindSignalingChannel(d *webrtc.DataChannel) {
	p.mux.Lock()
	p.dc = d
	p.mux.Unlock()

	d.OnOpen(func() {
		p.logger.Info("Signaling data channel open")

		p.mux.RLock()
		hasTracks := len(p.outTracks) > 0
		p.mux.RUnlock()

		if hasTracks {
			if err := p.Renegotiate(); err != nil {
				p.logger.Error("Failed to renegotiate on data channel open", slog.String("error", err.Error()))
			}
		}
	})

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
		if msg.SDP == nil {
			p.logger.Warn("Offer without SDP")
			return
		}

		answer, err := p.CreateAnswer(webrtc.SessionDescription{
			Type: webrtc.SDPTypeOffer,
			SDP:  *msg.SDP,
		})
		if err != nil {
			p.logger.Error("Failed to create answer", slog.String("error", err.Error()))
			return
		}

		answerMsg := entity.SignalMessage{
			Type:     entity.TypeAnswer,
			RoomID:   p.roomID,
			MemberID: p.id,
			SDP:      &answer.SDP,
		}

		if err := p.SendSignal(answerMsg); err != nil {
			p.logger.Error("Failed to send answer", slog.String("error", err.Error()))
		}

	case entity.TypeAnswer:
		if msg.SDP == nil {
			p.logger.Warn("Answer without SDP")
			return
		}

		p.negMux.Lock()
		err := p.SetAnswer(webrtc.SessionDescription{
			Type: webrtc.SDPTypeAnswer,
			SDP:  *msg.SDP,
		})
		pending := p.pendingNegotiation
		p.negMux.Unlock()

		if err != nil {
			p.logger.Error("Failed to set remote answer", slog.String("error", err.Error()))
			return
		}

		if pending {
			if err := p.Renegotiate(); err != nil {
				p.logger.Error("Failed to renegotiate", slog.String("error", err.Error()))
			}
		}

	case entity.TypeLeave:
		p.notifyClose()

	default:
		p.logger.Debug("Ignoring signaling message", slog.String("type", string(msg.Type)))
	}
}

func (p *Peer) SendSignal(msg entity.SignalMessage) error {
	p.mux.RLock()
	dc := p.dc
	p.mux.RUnlock()

	if dc == nil || dc.ReadyState() != webrtc.DataChannelStateOpen {
		return fmt.Errorf("peer %s: signaling data channel not open", p.id)
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal signal message: %w", err)
	}

	return dc.SendText(string(data))
}

func (p *Peer) RemoveOutboundTracks(trackIDs []string) error {
	removed, err := p.removeOutboundTracks(trackIDs)
	if removed > 0 {
		err = errors.Join(err, p.Renegotiate())
	}

	return err
}

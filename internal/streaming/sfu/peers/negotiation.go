package peers

import (
	"errors"
	"log/slog"

	"github.com/CMAK12/gonference/internal/streaming/entity"

	"github.com/pion/webrtc/v4"
)

func (c *connection) AddTrackAndRenegotiate(track *webrtc.TrackLocalStaticRTP) error {
	added, err := c.AddOutboundTrack(track)
	if err != nil {
		return err
	}

	if !added {
		c.logger.Warn("Track already exists", slog.String("track", track.ID()))
		return nil
	}

	return c.renegotiate()
}

func (c *connection) RemoveOutboundTracks(trackIDs []string) error {
	removed, err := c.removeOutboundTracks(trackIDs)
	if removed > 0 {
		err = errors.Join(err, c.renegotiate())
	}

	return err
}

func (c *connection) renegotiate() error {
	c.mux.RLock()
	dc := c.dc
	c.mux.RUnlock()

	if dc == nil || dc.ReadyState() != webrtc.DataChannelStateOpen {
		return nil
	}

	c.negMux.Lock()
	defer c.negMux.Unlock()

	if c.pc.SignalingState() != webrtc.SignalingStateStable {
		c.pendingNegotiation = true
		return nil
	}
	c.pendingNegotiation = false

	offer, err := c.CreateOffer()
	if err != nil {
		return err
	}

	msg := entity.SignalMessage{
		Type:     entity.TypeOffer,
		RoomID:   c.roomID,
		MemberID: c.id,
		SDP:      &offer.SDP,
	}

	return c.SendSignal(msg)
}

func (c *connection) onSignalingOpen() {
	c.logger.Info("Signaling data channel open")

	c.mux.RLock()
	hasTracks := len(c.outTracks) > 0
	c.mux.RUnlock()

	if !hasTracks {
		return
	}

	if err := c.renegotiate(); err != nil {
		c.logger.Error("Failed to renegotiate on data channel open", slog.String("error", err.Error()))
	}
}

func (c *connection) handleOffer(msg entity.SignalMessage) {
	if msg.SDP == nil {
		c.logger.Warn("Offer without SDP")
		return
	}

	answer, err := c.CreateAnswer(webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  *msg.SDP,
	})
	if err != nil {
		c.logger.Error("Failed to create answer", slog.String("error", err.Error()))
		return
	}

	answerMsg := entity.SignalMessage{
		Type:     entity.TypeAnswer,
		RoomID:   c.roomID,
		MemberID: c.id,
		SDP:      &answer.SDP,
	}

	if err := c.SendSignal(answerMsg); err != nil {
		c.logger.Error("Failed to send answer", slog.String("error", err.Error()))
	}
}

func (c *connection) handleAnswer(msg entity.SignalMessage) {
	if msg.SDP == nil {
		c.logger.Warn("Answer without SDP")
		return
	}

	c.negMux.Lock()
	err := c.SetAnswer(webrtc.SessionDescription{
		Type: webrtc.SDPTypeAnswer,
		SDP:  *msg.SDP,
	})
	pending := c.pendingNegotiation
	c.negMux.Unlock()

	if err != nil {
		c.logger.Error("Failed to set remote answer", slog.String("error", err.Error()))
		return
	}

	if pending {
		if err := c.renegotiate(); err != nil {
			c.logger.Error("Failed to renegotiate", slog.String("error", err.Error()))
		}
	}
}

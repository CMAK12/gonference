package peers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"sync"

	"github.com/CMAK12/gonference/internal/streaming/entity"

	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v4"
)

const signalingChannelLabel = "signaling"

type Callbacks struct {
	OnTrack       func(remote *webrtc.TrackRemote)
	OnClose       func()
	OnMemberLeave func(memberID string)
}

type connection struct {
	id        string
	roomID    string
	pc        *webrtc.PeerConnection
	dc        *webrtc.DataChannel
	callbacks Callbacks
	logger    *slog.Logger

	mux            sync.RWMutex
	inTracks       map[string]*webrtc.TrackRemote
	outTracks      map[string]*webrtc.RTPSender
	candidateQueue []webrtc.ICECandidateInit

	negMux             sync.Mutex
	pendingNegotiation bool

	closeOnce  sync.Once
	notifyOnce sync.Once
	closeErr   error
}

func newConnection(api *webrtc.API, roomID, id string, callbacks Callbacks) (*connection, error) {
	pc, err := api.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("new peer connection: %w", err)
	}

	c := &connection{
		id:        id,
		roomID:    roomID,
		pc:        pc,
		callbacks: callbacks,
		logger:    slog.Default().With("component", "peer", "id", id),
		inTracks:  make(map[string]*webrtc.TrackRemote),
		outTracks: make(map[string]*webrtc.RTPSender),
	}

	pc.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		c.addInboundTrack(track)
		c.callbacks.OnTrack(track)
	})

	pc.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		c.logger.Debug("Connection state change", slog.String("state", state.String()))

		switch state {
		case webrtc.PeerConnectionStateClosed, webrtc.PeerConnectionStateFailed:
			c.notifyClose()
		}
	})

	return c, nil
}

func (c *connection) ID() string {
	return c.id
}

func (c *connection) Logger() *slog.Logger {
	return c.logger
}

func (c *connection) AddICECandidate(ice webrtc.ICECandidateInit) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	if c.pc.RemoteDescription() == nil {
		c.candidateQueue = append(c.candidateQueue, ice)
		return nil
	}

	return c.pc.AddICECandidate(ice)
}

func (c *connection) CreateOffer() (webrtc.SessionDescription, error) {
	offer, err := c.pc.CreateOffer(nil)
	if err != nil {
		return webrtc.SessionDescription{}, err
	}

	if err = c.pc.SetLocalDescription(offer); err != nil {
		return webrtc.SessionDescription{}, err
	}

	<-webrtc.GatheringCompletePromise(c.pc)

	return *c.pc.LocalDescription(), nil
}

func (c *connection) CreateAnswer(offer webrtc.SessionDescription) (webrtc.SessionDescription, error) {
	if err := c.pc.SetRemoteDescription(offer); err != nil {
		return webrtc.SessionDescription{}, err
	}

	c.flushCandidateQueue()

	answer, err := c.pc.CreateAnswer(nil)
	if err != nil {
		return webrtc.SessionDescription{}, err
	}

	if err = c.pc.SetLocalDescription(answer); err != nil {
		return webrtc.SessionDescription{}, err
	}

	<-webrtc.GatheringCompletePromise(c.pc)

	return *c.pc.LocalDescription(), nil
}

func (c *connection) SetAnswer(offer webrtc.SessionDescription) error {
	return c.pc.SetRemoteDescription(offer)
}

func (c *connection) AddOutboundTrack(t *webrtc.TrackLocalStaticRTP) (bool, error) {
	c.mux.Lock()
	defer c.mux.Unlock()

	if _, exists := c.outTracks[t.ID()]; exists {
		return false, nil
	}

	sender, err := c.pc.AddTrack(t)
	if err != nil {
		return false, err
	}
	c.outTracks[t.ID()] = sender

	return true, nil
}

func (c *connection) SendSignal(msg entity.SignalMessage) error {
	c.mux.RLock()
	dc := c.dc
	c.mux.RUnlock()

	if dc == nil || dc.ReadyState() != webrtc.DataChannelStateOpen {
		return fmt.Errorf("peer %s: signaling data channel not open", c.id)
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal signal message: %w", err)
	}

	return dc.SendText(string(data))
}

func (c *connection) RequestKeyframe(ssrc uint32) {
	if err := c.pc.WriteRTCP([]rtcp.Packet{
		&rtcp.PictureLossIndication{
			MediaSSRC: ssrc,
		},
	}); err != nil {
		c.logger.Error("Failed to send PLI", slog.String("error", err.Error()))
	}

	c.logger.Info("Sent PLI", slog.Uint64("ssrc", uint64(ssrc)))
}

func (c *connection) removeOutboundTracks(ids []string) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}

	var (
		removed int
		errs    []error
	)

	c.mux.Lock()
	for _, id := range ids {
		track, ok := c.outTracks[id]
		if !ok {
			continue
		}
		delete(c.outTracks, id)

		if err := c.pc.RemoveTrack(track); err != nil {
			errs = append(errs, fmt.Errorf("remove track %s: %w", id, err))
			continue
		}
		removed++
	}
	c.mux.Unlock()

	return removed, errors.Join(errs...)
}

func (c *connection) flushCandidateQueue() {
	c.mux.Lock()
	defer c.mux.Unlock()
	for _, candidate := range c.candidateQueue {
		if err := c.pc.AddICECandidate(candidate); err != nil {
			c.logger.Error("Failed to add ICECandidate", slog.String("error", err.Error()))
		}
	}
	c.candidateQueue = nil
}

func (c *connection) addInboundTrack(track *webrtc.TrackRemote) {
	c.mux.Lock()
	defer c.mux.Unlock()

	c.inTracks[track.ID()] = track
}

func (c *connection) notifyClose() {
	c.notifyOnce.Do(c.callbacks.OnClose)
}

func (c *connection) Close() error {
	c.mux.Lock()
	dc := c.dc
	c.dc = nil
	c.mux.Unlock()

	c.negMux.Lock()
	c.pendingNegotiation = false
	c.negMux.Unlock()

	var errs []error
	if dc != nil {
		if err := dc.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
			errs = append(errs, fmt.Errorf("close signaling channel: %w", err))
		}
	}

	if err := c.close(); err != nil && !errors.Is(err, net.ErrClosed) {
		errs = append(errs, fmt.Errorf("close peer connection: %w", err))
	}

	return errors.Join(errs...)
}

func (c *connection) close() error {
	c.closeOnce.Do(func() {
		c.mux.Lock()
		clear(c.inTracks)
		clear(c.outTracks)
		c.candidateQueue = nil
		c.mux.Unlock()
		c.closeErr = c.pc.Close()
	})

	return c.closeErr
}

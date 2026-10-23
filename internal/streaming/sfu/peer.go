package sfu

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"sync"

	"github.com/CMAK12/gonference/internal/streaming/entity"

	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v3"
)

const signalingChannelLabel = "signaling"

type Peer struct {
	id string

	logger *slog.Logger
	conn   *webrtc.PeerConnection
	room   *Room

	negMux             sync.Mutex
	pendingNegotiation bool

	mux            sync.RWMutex
	dc             *webrtc.DataChannel
	inTracks       map[string]*webrtc.TrackRemote
	outTracks      map[string]*webrtc.RTPSender
	candidateQueue []webrtc.ICECandidateInit

	closeOnce sync.Once
	closeErr  error
}

func NewPeer(api *webrtc.API, room *Room, id string) (*Peer, error) {
	pc, err := api.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	peer := &Peer{
		id:        id,
		logger:    slog.Default().With("peer", id),
		conn:      pc,
		room:      room,
		inTracks:  make(map[string]*webrtc.TrackRemote),
		outTracks: make(map[string]*webrtc.RTPSender),
	}

	pc.OnDataChannel(func(d *webrtc.DataChannel) {
		if d.Label() != signalingChannelLabel {
			return
		}

		peer.bindSignalingChannel(d)
	})

	var cleanupOnce sync.Once
	cleanup := func() {
		peer.room.removePeer(peer)
	}

	pc.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		peer.logger.Info("Connection state change", slog.String("state", state.String()))

		switch state {
		case webrtc.PeerConnectionStateClosed,
			webrtc.PeerConnectionStateFailed,
			webrtc.PeerConnectionStateDisconnected:
			cleanupOnce.Do(cleanup)
		}
	})

	pc.OnTrack(func(remote *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		peer.addInboundTrack(remote)
		peer.room.addIncomingTrack(peer, remote)
	})

	_, err = pc.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo, webrtc.RTPTransceiverInit{
		Direction: webrtc.RTPTransceiverDirectionRecvonly,
	})
	if err != nil {
		return nil, err
	}

	return peer, nil
}

func (p *Peer) ID() string {
	return p.id
}

func (p *Peer) CreateAnswer(offer webrtc.SessionDescription) (webrtc.SessionDescription, error) {
	if err := p.conn.SetRemoteDescription(offer); err != nil {
		return webrtc.SessionDescription{}, err
	}

	p.flushCandidateQueue()

	answer, err := p.conn.CreateAnswer(nil)
	if err != nil {
		return webrtc.SessionDescription{}, err
	}

	if err = p.conn.SetLocalDescription(answer); err != nil {
		return webrtc.SessionDescription{}, err
	}

	<-webrtc.GatheringCompletePromise(p.conn)

	return *p.conn.LocalDescription(), nil
}

func (p *Peer) CreateOffer() (webrtc.SessionDescription, error) {
	offer, err := p.conn.CreateOffer(nil)
	if err != nil {
		return webrtc.SessionDescription{}, err
	}

	if err = p.conn.SetLocalDescription(offer); err != nil {
		return webrtc.SessionDescription{}, err
	}

	<-webrtc.GatheringCompletePromise(p.conn)

	return *p.conn.LocalDescription(), nil
}

func (p *Peer) ValidateAnswer(answer webrtc.SessionDescription) error {
	return p.conn.SetRemoteDescription(answer)
}

func (p *Peer) AddICECandidate(ci webrtc.ICECandidateInit) error {
	p.mux.Lock()
	defer p.mux.Unlock()

	if p.conn.RemoteDescription() == nil {
		p.candidateQueue = append(p.candidateQueue, ci)
		return nil
	}
	return p.conn.AddICECandidate(ci)
}

func (p *Peer) AddTrackAndRenegotiate(track *webrtc.TrackLocalStaticRTP) error {
	added, err := p.addOutboundTrack(track)
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

	if p.conn.SignalingState() != webrtc.SignalingStateStable {
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
		RoomID:   p.room.ID(),
		MemberID: p.id,
		SDP:      &offer.SDP,
	}

	return p.sendSignal(msg)
}

func (p *Peer) SendPLI(ssrc uint32) {
	if err := p.conn.WriteRTCP([]rtcp.Packet{
		&rtcp.PictureLossIndication{
			MediaSSRC: ssrc,
		},
	}); err != nil {
		p.logger.Error("Failed to send PLI", slog.String("error", err.Error()))
	}

	p.logger.Info(
		"Sent PLI",
		slog.String("peer", p.id),
		slog.Uint64("ssrc", uint64(ssrc)),
	)
}

func (p *Peer) Close() error {
	p.closeOnce.Do(func() {
		p.mux.Lock()
		dc := p.dc
		p.dc = nil
		p.candidateQueue = nil
		clear(p.inTracks)
		clear(p.outTracks)
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

		if err := p.conn.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
			errs = append(errs, fmt.Errorf("close peer connection: %w", err))
		}

		p.closeErr = errors.Join(errs...)
	})

	return p.closeErr
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
			RoomID:   p.room.ID(),
			MemberID: p.id,
			SDP:      &answer.SDP,
		}

		if err := p.sendSignal(answerMsg); err != nil {
			p.logger.Error("Failed to send answer", slog.String("error", err.Error()))
		}

	case entity.TypeAnswer:
		if msg.SDP == nil {
			p.logger.Warn("Answer without SDP")
			return
		}

		p.negMux.Lock()
		err := p.ValidateAnswer(webrtc.SessionDescription{
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
		p.room.removePeer(p)

	default:
		p.logger.Debug("Ignoring signaling message", slog.String("type", string(msg.Type)))
	}
}

func (p *Peer) sendSignal(msg entity.SignalMessage) error {
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

func (p *Peer) flushCandidateQueue() {
	p.mux.Lock()
	defer p.mux.Unlock()
	for _, c := range p.candidateQueue {
		if err := p.conn.AddICECandidate(c); err != nil {
			p.logger.Error("Failed to add ICECandidate", slog.String("error", err.Error()))
		}
	}
	p.candidateQueue = nil
}

func (p *Peer) addInboundTrack(track *webrtc.TrackRemote) {
	p.mux.Lock()
	defer p.mux.Unlock()

	p.inTracks[track.ID()] = track
}

func (p *Peer) addOutboundTrack(track *webrtc.TrackLocalStaticRTP) (bool, error) {
	p.mux.Lock()
	defer p.mux.Unlock()

	if _, exists := p.outTracks[track.ID()]; exists {
		return false, nil
	}

	sender, err := p.conn.AddTrack(track)
	if err != nil {
		return false, err
	}
	p.outTracks[track.ID()] = sender

	return true, nil
}

func (p *Peer) removeOutboundTracks(forwarders []*TrackForwarder) error {
	var (
		errs    []error
		removed int
	)

	p.mux.Lock()
	for _, forwarder := range forwarders {
		id := forwarder.remote.ID()
		sender, ok := p.outTracks[id]
		if !ok {
			continue
		}
		delete(p.outTracks, id)

		if err := p.conn.RemoveTrack(sender); err != nil {
			errs = append(errs, fmt.Errorf("remove track %s: %w", id, err))
			continue
		}
		removed++
	}
	p.mux.Unlock()

	if removed > 0 {
		errs = append(errs, p.Renegotiate())
	}

	return errors.Join(errs...)
}

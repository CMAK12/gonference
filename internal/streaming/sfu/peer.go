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

// signalingChannelLabel is the label of the DataChannel the client opens to
// exchange post-handshake signaling (renegotiation offers/answers, leave).
const signalingChannelLabel = "signaling"

type Peer struct {
	id string

	logger *slog.Logger
	conn   *webrtc.PeerConnection
	room   *Room

	mux            sync.RWMutex
	dc             *webrtc.DataChannel
	inTracks       map[string]*webrtc.TrackRemote
	outTracks      map[string]*webrtc.TrackLocalStaticRTP
	candidateQueue []webrtc.ICECandidateInit
}

// NewPeer creates a peer connection for the client's offer and returns the local
// answer. The answer is sent back over the bootstrap RPC; all further signaling
// flows over the client-created DataChannel.
func NewPeer(api *webrtc.API, room *Room, offer webrtc.SessionDescription, id string) (*Peer, webrtc.SessionDescription, error) {
	pc, err := api.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	})
	if err != nil {
		return nil, webrtc.SessionDescription{}, err
	}

	peer := &Peer{
		id:        id,
		logger:    slog.Default().With("peer", id),
		conn:      pc,
		room:      room,
		inTracks:  make(map[string]*webrtc.TrackRemote),
		outTracks: make(map[string]*webrtc.TrackLocalStaticRTP),
	}

	pc.OnDataChannel(func(d *webrtc.DataChannel) {
		if d.Label() != signalingChannelLabel {
			return
		}

		peer.mux.Lock()
		peer.dc = d
		peer.mux.Unlock()

		d.OnOpen(func() {
			peer.logger.Info("Signaling data channel open")

			peer.mux.RLock()
			hasTracks := len(peer.outTracks) > 0
			peer.mux.RUnlock()

			// Tracks added before the channel opened (existing room media)
			// are negotiated now that we have a channel to deliver the offer.
			if hasTracks {
				if err := peer.Renegotiate(); err != nil {
					peer.logger.Error("Failed to renegotiate on data channel open", slog.String("error", err.Error()))
				}
			}
		})

		d.OnMessage(func(raw webrtc.DataChannelMessage) {
			var msg entity.SignalMessage
			if err := json.Unmarshal(raw.Data, &msg); err != nil {
				peer.logger.Error("Failed to unmarshal signaling message", slog.String("error", err.Error()))
				return
			}

			peer.handleSignal(msg)
		})
	})

	var cleanupOnce sync.Once
	cleanup := func() {
		peer.room.RemovePeer(peer.id)
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
		return nil, webrtc.SessionDescription{}, err
	}

	answer, err := peer.CreateAnswer(offer)
	if err != nil {
		return nil, webrtc.SessionDescription{}, err
	}

	return peer, answer, nil
}

func (p *Peer) ID() string {
	return p.id
}

func (p *Peer) Close() error {
	p.mux.Lock()
	dc := p.dc
	p.dc = nil
	clear(p.inTracks)
	clear(p.outTracks)
	p.mux.Unlock()

	if dc != nil {
		_ = dc.Close()
	}

	if err := p.conn.Close(); err != nil {
		if errors.Is(err, net.ErrClosed) {
			p.logger.Debug("Peer connection already closed")
		} else {
			p.logger.Error("Failed to close peer connection", slog.String("error", err.Error()))
		}
	}

	return nil
}

func (p *Peer) SendPLI(ssrc uint32) {
	p.logger.Info(
		"Send PLI",
		slog.String("peer", p.id),
		slog.Uint64("ssrc", uint64(ssrc)),
	)

	_ = p.conn.WriteRTCP([]rtcp.Packet{
		&rtcp.PictureLossIndication{
			MediaSSRC: ssrc,
		},
	})
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
	p.mux.Lock()

	if _, exists := p.outTracks[track.ID()]; exists {
		p.mux.Unlock()
		p.logger.Warn("Track already exists", slog.String("track", track.ID()))
		return nil
	}

	p.outTracks[track.ID()] = track
	p.mux.Unlock()

	if _, err := p.conn.AddTrack(track); err != nil {
		return err
	}

	return p.Renegotiate()
}

// Renegotiate creates a fresh offer and sends it over the signaling DataChannel.
// If the channel is not open yet, it is a no-op: the OnOpen handler renegotiates
// once the channel becomes available, capturing all tracks added in the meantime.
func (p *Peer) Renegotiate() error {
	p.mux.RLock()
	dc := p.dc
	p.mux.RUnlock()

	if dc == nil || dc.ReadyState() != webrtc.DataChannelStateOpen {
		return nil
	}

	offer, err := p.conn.CreateOffer(nil)
	if err != nil {
		return err
	}

	if err = p.conn.SetLocalDescription(offer); err != nil {
		return err
	}

	<-webrtc.GatheringCompletePromise(p.conn)

	msg := entity.SignalMessage{
		Type:     entity.TypeOffer,
		RoomID:   p.room.ID(),
		MemberID: p.id,
		SDP:      &p.conn.LocalDescription().SDP,
	}

	return p.sendSignal(msg)
}

// sendSignal serializes a signaling message as JSON and writes it to the client
// over the signaling DataChannel.
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

// handleSignal dispatches an inbound signaling message received on the DataChannel.
func (p *Peer) handleSignal(msg entity.SignalMessage) {
	switch msg.Type {
	case entity.TypeAnswer:
		if msg.SDP == nil {
			p.logger.Warn("Answer without SDP")
			return
		}

		if err := p.ValidateAnswer(webrtc.SessionDescription{
			Type: webrtc.SDPTypeAnswer,
			SDP:  *msg.SDP,
		}); err != nil {
			p.logger.Error("Failed to set remote answer", slog.String("error", err.Error()))
		}

	case entity.TypeLeave:
		p.room.RemovePeer(p.id)

	default:
		p.logger.Debug("Ignoring signaling message", slog.String("type", string(msg.Type)))
	}
}

func (p *Peer) flushCandidateQueue() {
	p.mux.Lock()
	defer p.mux.Unlock()
	for _, c := range p.candidateQueue {
		_ = p.conn.AddICECandidate(c)
	}
	p.candidateQueue = nil
}

func (p *Peer) addInboundTrack(track *webrtc.TrackRemote) {
	p.mux.Lock()
	defer p.mux.Unlock()

	p.inTracks[track.ID()] = track
}

func (p *Peer) addOutboundTrack(track *webrtc.TrackLocalStaticRTP) {
	p.mux.Lock()
	defer p.mux.Unlock()

	p.outTracks[track.ID()] = track
}

package sfu

import (
	"errors"
	"gonference/internal/entity"
	"log/slog"
	"net"
	"sync"

	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v3"
)

type Signaling interface {
	WriteMessage(msg entity.Message) error
	Close() error
}

type Peer struct {
	id string

	logger    *slog.Logger
	conn      *webrtc.PeerConnection
	room      *Room
	signaling Signaling

	mux            sync.RWMutex
	inTracks       map[string]*webrtc.TrackRemote
	outTracks      map[string]*webrtc.TrackLocalStaticRTP
	candidateQueue []webrtc.ICECandidateInit
}

func NewPeer(api *webrtc.API, signal Signaling, room *Room, offer webrtc.SessionDescription, id string) (*Peer, error) {
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
		signaling: signal,
		inTracks:  make(map[string]*webrtc.TrackRemote),
		outTracks: make(map[string]*webrtc.TrackLocalStaticRTP),
	}

	pc.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c == nil {
			return
		}

		candidate := c.ToJSON()

		msg := entity.Message{
			Type:      entity.TypeCandidate,
			RoomID:    room.ID(),
			MemberID:  id,
			Candidate: &candidate,
		}

		if err := signal.WriteMessage(msg); err != nil {
			peer.logger.Error("Failed to send ICE candidate", slog.String("error", err.Error()))
		}
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
		return nil, err
	}

	_, err = pc.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio, webrtc.RTPTransceiverInit{
		Direction: webrtc.RTPTransceiverDirectionRecvonly,
	})
	if err != nil {
		return nil, err
	}

	if err := peer.SendAnswer(offer); err != nil {
		return nil, err
	}

	return peer, err
}

func (p *Peer) ID() string {
	return p.id
}

func (p *Peer) Close() error {
	clear(p.inTracks)
	clear(p.outTracks)

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

	return answer, nil
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

func (p *Peer) Renegotiate() error {
	offer, err := p.conn.CreateOffer(nil)
	if err != nil {
		return err
	}

	if err = p.conn.SetLocalDescription(offer); err != nil {
		return err
	}

	<-webrtc.GatheringCompletePromise(p.conn)

	msg := entity.Message{
		Type:     entity.TypeOffer,
		RoomID:   p.room.ID(),
		MemberID: p.id,
		SDP:      &p.conn.LocalDescription().SDP,
	}

	return p.signaling.WriteMessage(msg)
}

func (p *Peer) SendAnswer(offer webrtc.SessionDescription) error {
	if err := p.conn.SetRemoteDescription(offer); err != nil {
		return err
	}

	p.flushCandidateQueue()

	answer, err := p.conn.CreateAnswer(nil)
	if err != nil {
		return err
	}

	if err = p.conn.SetLocalDescription(answer); err != nil {
		return err
	}

	<-webrtc.GatheringCompletePromise(p.conn)

	msg := entity.Message{
		Type:     entity.TypeAnswer,
		RoomID:   p.room.ID(),
		MemberID: p.id,
		SDP:      &p.conn.LocalDescription().SDP,
	}

	return p.signaling.WriteMessage(msg)
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

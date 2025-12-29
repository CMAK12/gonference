package sfu

import (
	"encoding/json"
	"log/slog"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v3"
)

type Signaling interface {
	WriteMessage(msgType int, payload []byte) error
}

type Peer struct {
	ID string

	logger *slog.Logger
	conn   *webrtc.PeerConnection
	room   *Room
	signal Signaling

	renegotiationMux sync.Mutex
	mux              sync.RWMutex
	inTracks         map[string]*webrtc.TrackRemote
	outTracks        map[string]*webrtc.TrackLocalStaticRTP
	candidateQueue   []webrtc.ICECandidateInit
}

func NewPeer(api *webrtc.API, signal Signaling, room *Room, offer webrtc.SessionDescription, id string) (*Peer, error) {
	pc, err := api.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
			{URLs: []string{"turn:turn.example.com:3478"}, Username: "TURN_USER", Credential: "TURN_PASS"},
		},
	})

	peer := &Peer{
		ID:        id,
		logger:    slog.Default().With("peer", id),
		conn:      pc,
		room:      room,
		signal:    signal,
		inTracks:  make(map[string]*webrtc.TrackRemote),
		outTracks: make(map[string]*webrtc.TrackLocalStaticRTP),
	}

	pc.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c == nil {
			return
		}

		msg, err := json.Marshal(map[string]any{
			"type":      "candidate",
			"roomId":    room.ID(),
			"memberId":  id,
			"candidate": c.ToJSON(),
		})
		if err != nil {
			peer.logger.Error("Failed to marshal ICE candidate", slog.String("error", err.Error()))
			return
		}

		peer.signal.WriteMessage(websocket.TextMessage, msg)
	})

	pc.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		peer.logger.Info("Connection state change", slog.String("state", state.String()))
	})

	pc.OnTrack(func(remote *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		peer.inTracks[remote.ID()] = remote
		peer.room.addIncomingTrack(peer, remote)
	})

	pc.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo, webrtc.RTPTransceiverInit{
		Direction: webrtc.RTPTransceiverDirectionRecvonly,
	})

	if err := peer.SendAnswer(offer); err != nil {
		return nil, err
	}

	return peer, err
}

func (p *Peer) SendPLI(ssrc uint32) {
	p.logger.Info(
		"Send PLI",
		slog.String("peer", p.ID),
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
	p.renegotiationMux.Lock()
	defer p.renegotiationMux.Unlock()

	if _, err := p.conn.AddTrack(track); err != nil {
		return err
	}

	offer, err := p.conn.CreateOffer(nil)
	if err != nil {
		return err
	}

	if err = p.conn.SetLocalDescription(offer); err != nil {
		return err
	}

	<-webrtc.GatheringCompletePromise(p.conn)

	msg, err := json.Marshal(map[string]any{
		"type":     "offer",
		"roomId":   p.room.ID(),
		"memberId": p.ID,
		"sdp":      p.conn.LocalDescription().SDP,
	})
	if err != nil {
		return err
	}

	return p.signal.WriteMessage(websocket.TextMessage, msg)
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

	msg, err := json.Marshal(map[string]any{
		"type":     "offer",
		"roomId":   p.room.ID(),
		"memberId": p.ID,
		"sdp":      p.conn.LocalDescription().SDP,
	})
	if err != nil {
		return err
	}

	return p.signal.WriteMessage(websocket.TextMessage, msg)
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

	msg, err := json.Marshal(map[string]any{
		"type":     "answer",
		"roomId":   p.room.ID(),
		"memberId": p.ID,
		"sdp":      p.conn.LocalDescription().SDP,
	})
	if err != nil {
		return err
	}

	return p.signal.WriteMessage(websocket.TextMessage, msg)
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

package sfu

import (
	"encoding/json"
	"log/slog"

	"github.com/gorilla/websocket"
	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v3"
)

type Peer struct {
	id string

	conn   *webrtc.PeerConnection
	ws     Signaling
	logger *slog.Logger

	inTracks  map[string]*webrtc.TrackRemote
	outTracks map[string]*webrtc.TrackLocalStaticRTP

	room *Room
}

func NewPeer(api *webrtc.API, ws Signaling, id string, room *Room) (*Peer, error) {
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
		conn:      pc,
		ws:        ws,
		room:      room,
		logger:    slog.Default().With("component", "peer"),
		inTracks:  make(map[string]*webrtc.TrackRemote),
		outTracks: make(map[string]*webrtc.TrackLocalStaticRTP),
	}

	pc.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		peer.logger.Info(
			"Peer connection state changed",
			slog.String("state", state.String()),
		)

		switch state {
		case webrtc.PeerConnectionStateConnected:
			peer.logger.Info(
				"Peer connection connected",
				slog.String("peer", peer.id),
			)

		case webrtc.PeerConnectionStateDisconnected,
			webrtc.PeerConnectionStateClosed,
			webrtc.PeerConnectionStateFailed:
			peer.logger.Info(
				"Peer connection disconnected",
				slog.String("peer", peer.id),
			)
		}
	})

	pc.OnTrack(func(remote *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		peer.inTracks[remote.ID()] = remote
		room.AddIncomingTrack(peer, remote)
	})

	if _, err := pc.AddTransceiverFromKind(
		webrtc.RTPCodecTypeVideo,
		webrtc.RTPTransceiverInit{
			Direction: webrtc.RTPTransceiverDirectionRecvonly,
		},
	); err != nil {
		return nil, err
	}

	return peer, nil
}

func (p *Peer) SetRemoteDescription(sdp webrtc.SessionDescription) error {
	return p.conn.SetRemoteDescription(sdp)
}

func (p *Peer) CreateAnswer(offer webrtc.SessionDescription) (*webrtc.SessionDescription, error) {
	if err := p.conn.SetRemoteDescription(offer); err != nil {
		return nil, err
	}

	answer, err := p.conn.CreateAnswer(nil)
	if err != nil {
		return nil, err
	}

	if err := p.conn.SetLocalDescription(answer); err != nil {
		return nil, err
	}

	<-webrtc.GatheringCompletePromise(p.conn)

	return p.conn.LocalDescription(), nil
}

func (p *Peer) Renegotiate() error {
	offer, err := p.conn.CreateOffer(nil)
	if err != nil {
		return err
	}

	if err = p.conn.SetLocalDescription(offer); err != nil {
		return err
	}

	// Wait for ICE gathering to complete so SDP contains candidates
	<-webrtc.GatheringCompletePromise(p.conn)

	// Send a JSON message over the signaling channel instead of raw SDP
	msg := map[string]string{
		"type":     "offer",
		"roomId":   p.room.id,
		"memberId": p.id,
		"sdp":      p.conn.LocalDescription().SDP,
	}
	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return p.ws.WriteMessage(websocket.TextMessage, b)
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

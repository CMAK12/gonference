package sfu

import (
	"errors"
	"log/slog"
	"sync"

	"github.com/CMAK12/gonference/internal/streaming/sfu/peers"
	"github.com/CMAK12/gonference/internal/streaming/sfu/track"

	"github.com/pion/webrtc/v4"
)

type (
	Peer interface {
		ID() string
		Logger() *slog.Logger
		AddTrackAndRenegotiate(track *webrtc.TrackLocalStaticRTP) error
		AddOutboundTrack(track *webrtc.TrackLocalStaticRTP) (bool, error)
		RemoveOutboundTracks(trackIDs []string) error
		NotifyLeave(memberID string) error
		Close() error
	}

	Forwarder interface {
		PublisherID() string
		AddPeer(id string) (*webrtc.TrackLocalStaticRTP, error)
		RemovePeer(id string)
		Close()
	}
)

type Room struct {
	id string

	mux        sync.RWMutex
	peers      map[string]Peer
	forwarders map[string]Forwarder

	api    *webrtc.API
	logger *slog.Logger

	onEmpty func()
}

func NewRoom(api *webrtc.API, id string, onEmpty func()) *Room {
	return &Room{
		id:         id,
		api:        api,
		logger:     slog.Default().With("room_id", id),
		peers:      make(map[string]Peer),
		forwarders: make(map[string]Forwarder),

		onEmpty: onEmpty,
	}
}

func (r *Room) ID() string {
	return r.id
}

func (r *Room) GetPeer(id string) (Peer, bool) {
	r.mux.RLock()
	defer r.mux.RUnlock()

	peer, ok := r.peers[id]
	return peer, ok
}

func (r *Room) AddPeer(offer webrtc.SessionDescription, id string) (webrtc.SessionDescription, Peer, error) {
	var peer *peers.Peer
	peer, err := peers.NewPeer(r.api, r.id, id, peers.Callbacks{
		OnTrack: func(remote *webrtc.TrackRemote) {
			r.addIncomingTrack(id, peer.RequestKeyframe, remote)
		},
		OnClose: func() {
			r.removePeer(peer)
		},
	})
	if err != nil {
		return webrtc.SessionDescription{}, nil, err
	}

	answer, err := peer.CreateAnswer(offer)
	if err != nil {
		return webrtc.SessionDescription{}, nil, errors.Join(err, peer.Close())
	}

	r.attachPeer(peer)

	return answer, peer, nil
}

func (r *Room) RemovePeer(id string) {
	peer, ok := r.GetPeer(id)
	if !ok {
		return
	}

	r.removePeer(peer)
}

func (r *Room) Close() {
	r.mux.Lock()
	defer r.mux.Unlock()

	for _, peer := range r.peers {
		if err := peer.Close(); err != nil {
			peer.Logger().Error("Failed to close peer", slog.String("error", err.Error()))
		}
	}

	for _, forwarder := range r.forwarders {
		forwarder.Close()
	}
}

func (r *Room) attachPeer(peer Peer) {
	id := peer.ID()

	r.mux.Lock()
	replaced := r.peers[id]
	r.peers[id] = peer

	var detached detachedPeer
	if replaced != nil {
		detached = r.detachPeerLocked(id)
	}

	forwarders := make([]Forwarder, 0, len(r.forwarders))
	for _, f := range r.forwarders {
		if f.PublisherID() != id {
			forwarders = append(forwarders, f)
		}
	}
	r.mux.Unlock()

	if replaced != nil {
		r.releasePeer(replaced, detached)
	}

	for _, forwarder := range forwarders {
		local, err := forwarder.AddPeer(id)
		if err != nil {
			peer.Logger().Error("Failed to add peer to forwarder", slog.String("error", err.Error()))
			continue
		}

		if _, err := peer.AddOutboundTrack(local); err != nil {
			peer.Logger().Error("Failed to add track to peer connection", slog.String("error", err.Error()))
		}
	}
}

func (r *Room) addIncomingTrack(id string, sendPLI func(ssrc uint32), remote *webrtc.TrackRemote) {
	r.mux.Lock()

	forwarder := track.NewForwarder(id, remote, sendPLI)
	r.forwarders[remote.ID()] = forwarder

	subscribers := make(map[string]Peer, len(r.peers))
	for peerID, peer := range r.peers {
		if peerID != id {
			subscribers[peerID] = peer
		}
	}
	r.mux.Unlock()

	forwarder.Start()

	for peerID, peer := range subscribers {
		local, err := forwarder.AddPeer(peerID)
		if err != nil {
			r.logger.Error("Failed to add peer to forwarder",
				slog.String("peerId", peerID),
				slog.String("error", err.Error()))
			continue
		}

		if err := peer.AddTrackAndRenegotiate(local); err != nil {
			r.logger.Error("Failed to renegotiate",
				slog.String("peerId", peerID),
				slog.String("error", err.Error()))
			continue
		}
	}
}

func (r *Room) removePeer(peer Peer) {
	id := peer.ID()

	r.mux.Lock()
	if r.peers[id] != peer {
		r.mux.Unlock()
		return
	}
	delete(r.peers, id)
	detached := r.detachPeerLocked(id)
	isEmpty := len(r.peers) == 0
	r.mux.Unlock()

	r.releasePeer(peer, detached)
	r.notifyLeave(peer)

	if isEmpty {
		r.onEmpty()
	}
}

type detachedPeer struct {
	forwarders []Forwarder
	trackIDs   []string
	others     []Peer
}

func (r *Room) detachPeerLocked(id string) detachedPeer {
	var d detachedPeer
	for trackID, forwarder := range r.forwarders {
		if forwarder.PublisherID() == id {
			d.forwarders = append(d.forwarders, forwarder)
			d.trackIDs = append(d.trackIDs, trackID)
			delete(r.forwarders, trackID)
			continue
		}
		forwarder.RemovePeer(id)
	}

	d.others = make([]Peer, 0, len(r.peers))
	for peerID, p := range r.peers {
		if peerID != id {
			d.others = append(d.others, p)
		}
	}

	return d
}

func (r *Room) releasePeer(peer Peer, d detachedPeer) {
	for _, forwarder := range d.forwarders {
		forwarder.Close()
	}

	if err := peer.Close(); err != nil {
		peer.Logger().Error("Failed to close peer", slog.String("error", err.Error()))
	}

	if len(d.trackIDs) == 0 {
		return
	}

	for _, p := range d.others {
		if err := p.RemoveOutboundTracks(d.trackIDs); err != nil {
			p.Logger().Error("Failed to remove outbound tracks", slog.String("error", err.Error()))
		}
	}
}

func (r *Room) notifyLeave(left Peer) {
	r.mux.RLock()
	recipients := make([]Peer, 0, len(r.peers))
	for _, p := range r.peers {
		if p != left {
			recipients = append(recipients, p)
		}
	}
	r.mux.RUnlock()

	for _, peer := range recipients {
		if err := peer.NotifyLeave(left.ID()); err != nil {
			peer.Logger().Error("Failed to notify peer leave", slog.String("error", err.Error()))
		}
	}
}

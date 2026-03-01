package sfu

import (
	"sync"

	"github.com/pion/webrtc/v3"
)

type TrackForwarder struct {
	peer   *Peer
	remote *webrtc.TrackRemote

	mux    sync.RWMutex
	locals map[string]*webrtc.TrackLocalStaticRTP

	closed chan struct{}
}

func NewTrackForwarder(peer *Peer, remote *webrtc.TrackRemote) *TrackForwarder {
	return &TrackForwarder{
		peer:   peer,
		remote: remote,
		locals: make(map[string]*webrtc.TrackLocalStaticRTP),
		closed: make(chan struct{}),
	}
}

func (tf *TrackForwarder) AddPeer(id string) (*webrtc.TrackLocalStaticRTP, error) {
	local, err := webrtc.NewTrackLocalStaticRTP(
		tf.remote.Codec().RTPCodecCapability,
		tf.remote.ID(),
		tf.remote.StreamID(),
	)
	if err != nil {
		return nil, err
	}

	tf.mux.Lock()
	tf.locals[id] = local
	tf.mux.Unlock()

	go tf.peer.SendPLI(uint32(tf.remote.SSRC()))

	return local, nil
}

func (tf *TrackForwarder) RemovePeer(id string) {
	tf.mux.Lock()
	delete(tf.locals, id)
	tf.mux.Unlock()
}

func (tf *TrackForwarder) Start() {
	go func() {
		buf := make([]byte, 1500)

		for {
			select {
			case <-tf.closed:
				return
			default:
				n, _, err := tf.remote.Read(buf)
				if err != nil {
					return
				}

				tf.mux.RLock()
				for _, local := range tf.locals {
					if _, err = local.Write(buf[:n]); err != nil {
						continue
					}
				}
				tf.mux.RUnlock()
			}
		}
	}()
}

func (tf *TrackForwarder) Close() {
	close(tf.closed)
}

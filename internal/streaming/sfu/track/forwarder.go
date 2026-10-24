package track

import (
	"sync"

	"github.com/pion/webrtc/v4"
)

type Forwarder struct {
	publisherID     string
	remote          *webrtc.TrackRemote
	requestKeyframe func(ssrc uint32)

	mux    sync.RWMutex
	locals map[string]*webrtc.TrackLocalStaticRTP

	closed    chan struct{}
	closeOnce sync.Once
}

func NewForwarder(publisherID string, remote *webrtc.TrackRemote, requestKeyframe func(ssrc uint32)) *Forwarder {
	return &Forwarder{
		publisherID:     publisherID,
		remote:          remote,
		requestKeyframe: requestKeyframe,
		locals:          make(map[string]*webrtc.TrackLocalStaticRTP),
		closed:          make(chan struct{}),
	}
}

func (tf *Forwarder) PublisherID() string {
	return tf.publisherID
}

func (tf *Forwarder) TrackID() string {
	return tf.remote.ID()
}

func (tf *Forwarder) AddPeer(id string) (*webrtc.TrackLocalStaticRTP, error) {
	local, err := webrtc.NewTrackLocalStaticRTP(
		tf.remote.Codec().RTPCodecCapability,
		tf.remote.ID(),
		tf.publisherID,
	)
	if err != nil {
		return nil, err
	}

	tf.mux.Lock()
	tf.locals[id] = local
	tf.mux.Unlock()

	go tf.requestKeyframe(uint32(tf.remote.SSRC()))

	return local, nil
}

func (tf *Forwarder) RemovePeer(id string) {
	tf.mux.Lock()
	delete(tf.locals, id)
	tf.mux.Unlock()
}

func (tf *Forwarder) Start() {
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

func (tf *Forwarder) Close() {
	tf.closeOnce.Do(func() {
		close(tf.closed)
	})
}

package codecs

import "github.com/pion/webrtc/v4"

const OPUS webrtc.PayloadType = 109

var Audio = []webrtc.RTPCodecParameters{
	{
		RTPCodecCapability: webrtc.RTPCodecCapability{
			MimeType:  webrtc.MimeTypeOpus,
			ClockRate: 48000,
		},
		PayloadType: OPUS,
	},
}

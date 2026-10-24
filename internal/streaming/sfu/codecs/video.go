package codecs

import "github.com/pion/webrtc/v4"

const (
	VP8  webrtc.PayloadType = 120
	H264 webrtc.PayloadType = 130
)

var Video = []webrtc.RTPCodecParameters{
	//{
	//	RTPCodecCapability: webrtc.RTPCodecCapability{
	//		MimeType:  webrtc.MimeTypeAV1,
	//		ClockRate: 90000,
	//	},
	//	PayloadType: 96,
	//},
	//{
	//	RTPCodecCapability: webrtc.RTPCodecCapability{
	//		MimeType:    webrtc.MimeTypeVP9,
	//		ClockRate:   90000,
	//		SDPFmtpLine: "profile-id=0",
	//	},
	//	PayloadType: 97,
	//},
	//{
	//	RTPCodecCapability: webrtc.RTPCodecCapability{
	//		MimeType:    webrtc.MimeTypeVP9,
	//		ClockRate:   90000,
	//		SDPFmtpLine: "profile-id=1",
	//	},
	//	PayloadType: 121,
	//},
	{
		RTPCodecCapability: webrtc.RTPCodecCapability{
			MimeType:  webrtc.MimeTypeVP8,
			ClockRate: 90000,
		},
		PayloadType: VP8,
	},
	//{
	//	RTPCodecCapability: webrtc.RTPCodecCapability{
	//		MimeType:    webrtc.MimeTypeH264,
	//		ClockRate:   90000,
	//		SDPFmtpLine: "level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=42001f",
	//	},
	//	PayloadType: 97,
	//},
	{
		RTPCodecCapability: webrtc.RTPCodecCapability{
			MimeType:    webrtc.MimeTypeH264,
			ClockRate:   90000,
			SDPFmtpLine: "level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=42e01f",
		},
		PayloadType: H264,
	},
}

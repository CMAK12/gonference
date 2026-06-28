export type SignalType = 'offer' | 'answer' | 'candidate' | 'leave'

// SignalMessage mirrors the streaming service's entity.SignalMessage. The
// gateway only carries the initial offer/answer over HTTP (WHIP); these
// messages are exchanged afterwards over the peer's WebRTC DataChannel.
export interface SignalMessage {
  type: SignalType
  roomId: string
  memberId: string
  sdp?: string
  candidate?: RTCIceCandidateInit
}

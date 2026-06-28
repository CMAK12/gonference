import { markRaw, reactive, ref, shallowRef } from 'vue'

import { postWhipOffer } from '../lib/whip'
import type { SignalMessage } from '../types'

const ICE_SERVERS: RTCIceServer[] = [{ urls: 'stun:stun.l.google.com:19302' }]

// signalingChannelLabel must match the label the SFU listens for
// (sfu.signalingChannelLabel on the Go side).
const signalingChannelLabel = 'signaling'

export interface RemotePeer {
  memberId: string
  stream: MediaStream
}

// useConference drives a single membership in a conference room:
//   1. capture local media and create the publishing PeerConnection
//   2. bootstrap the session over WHIP (HTTP offer -> answer)
//   3. handle SFU-initiated renegotiation and leave events over the DataChannel
//
// There is no WebSocket: the only signaling transports are the one-shot WHIP
// POST and the WebRTC DataChannel.
export function useConference() {
  const roomId = ref('')
  const memberId = ref('')
  const connected = ref(false)
  const connecting = ref(false)
  const error = ref<string | null>(null)

  const localStream = shallowRef<MediaStream | null>(null)
  // Keyed by the publisher's member id, which equals the forwarded stream id
  // the SFU tags each track with (TrackForwarder uses the publisher peer id
  // as the local track's StreamID).
  const remotePeers = reactive<Record<string, RemotePeer>>({})

  let pc: RTCPeerConnection | null = null
  let dc: RTCDataChannel | null = null

  async function join(room: string): Promise<void> {
    if (connected.value || connecting.value) return

    error.value = null
    connecting.value = true

    try {
      roomId.value = room
      memberId.value = crypto.randomUUID()

      const stream = await navigator.mediaDevices.getUserMedia({ video: true, audio: false })
      localStream.value = markRaw(stream)

      pc = new RTCPeerConnection({ iceServers: ICE_SERVERS })

      // The DataChannel must be created before the offer so its m-line is part
      // of the negotiated session; the SFU pushes renegotiation offers here.
      dc = pc.createDataChannel(signalingChannelLabel)
      dc.onmessage = handleSignal

      pc.ontrack = (event) => {
        const remote = event.streams[0]
        if (!remote) return
        remotePeers[remote.id] = { memberId: remote.id, stream: markRaw(remote) }
      }

      pc.onconnectionstatechange = () => {
        const state = pc?.connectionState
        if (state === 'failed' || state === 'closed' || state === 'disconnected') {
          if (connected.value) leave()
        }
      }

      stream.getTracks().forEach((track) => pc!.addTrack(track, stream))

      const offer = await pc.createOffer()
      await pc.setLocalDescription(offer)

      // No trickle ICE channel exists, so gather fully before sending.
      await waitForIceGathering(pc)

      const answer = await postWhipOffer(roomId.value, memberId.value, pc.localDescription!.sdp)
      await pc.setRemoteDescription({ type: 'answer', sdp: answer })

      connected.value = true
    } catch (e) {
      error.value = e instanceof Error ? e.message : String(e)
      cleanup()
    } finally {
      connecting.value = false
    }
  }

  async function handleSignal(event: MessageEvent): Promise<void> {
    if (!pc) return

    let msg: SignalMessage
    try {
      msg = JSON.parse(event.data) as SignalMessage
    } catch {
      return
    }

    switch (msg.type) {
      case 'offer': {
        if (!msg.sdp) return
        await pc.setRemoteDescription({ type: 'offer', sdp: msg.sdp })
        const answer = await pc.createAnswer()
        await pc.setLocalDescription(answer)
        send({
          type: 'answer',
          roomId: roomId.value,
          memberId: memberId.value,
          sdp: pc.localDescription!.sdp,
        })
        break
      }
      case 'leave': {
        delete remotePeers[msg.memberId]
        break
      }
      default:
        break
    }
  }

  function send(msg: SignalMessage): void {
    if (dc && dc.readyState === 'open') {
      dc.send(JSON.stringify(msg))
    }
  }

  function leave(): void {
    send({ type: 'leave', roomId: roomId.value, memberId: memberId.value })
    cleanup()
    connected.value = false
  }

  function cleanup(): void {
    try {
      dc?.close()
    } catch {
      // already closed
    }
    try {
      pc?.close()
    } catch {
      // already closed
    }
    dc = null
    pc = null

    localStream.value?.getTracks().forEach((track) => track.stop())
    localStream.value = null

    for (const key of Object.keys(remotePeers)) {
      delete remotePeers[key]
    }
  }

  return {
    roomId,
    memberId,
    connected,
    connecting,
    error,
    localStream,
    remotePeers,
    join,
    leave,
  }
}

function waitForIceGathering(connection: RTCPeerConnection): Promise<void> {
  if (connection.iceGatheringState === 'complete') return Promise.resolve()

  return new Promise((resolve) => {
    const check = () => {
      if (connection.iceGatheringState === 'complete') {
        connection.removeEventListener('icegatheringstatechange', check)
        resolve()
      }
    }
    connection.addEventListener('icegatheringstatechange', check)
  })
}

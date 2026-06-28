const GATEWAY_URL = import.meta.env.VITE_GATEWAY_URL ?? ''

// postWhipOffer performs the WHIP bootstrap: it POSTs the local SDP offer as
// `application/sdp` and returns the SFU's answer SDP. Room and member are
// passed as query parameters since the body is pure SDP.
export async function postWhipOffer(
  roomId: string,
  memberId: string,
  offerSdp: string,
): Promise<string> {
  const url = `${GATEWAY_URL}/whip?room=${encodeURIComponent(roomId)}&member=${encodeURIComponent(memberId)}`

  const res = await fetch(url, {
    method: 'POST',
    headers: { 'Content-Type': 'application/sdp' },
    body: offerSdp,
  })

  if (!res.ok) {
    const detail = await res.text().catch(() => '')
    throw new Error(`WHIP request failed (${res.status}): ${detail || res.statusText}`)
  }

  return res.text()
}

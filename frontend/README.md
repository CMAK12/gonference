# gonference frontend

Vue 3 + Vite + TypeScript client for the gonference SFU.

## Signaling model

There is **no WebSocket**. A member's session is established with a single
WHIP request and kept current over a WebRTC DataChannel:

1. Capture local media and build an `RTCPeerConnection` with a `signaling`
   DataChannel.
2. Create an offer, gather ICE fully, and `POST` it to the gateway as
   `application/sdp` (WHIP): `POST /whip?room=<id>&member=<id>`. The gateway
   forwards it to the streaming SFU over gRPC and returns the answer SDP.
3. As other members join or leave, the SFU pushes renegotiation offers and
   `leave` notices over the DataChannel; the client answers in kind.

## Develop

```bash
npm install
npm run dev      # http://localhost:5173, proxies /whip -> http://localhost:8080
```

Run the gateway (`:8080`) and streaming (`:9090`) services alongside it.

## Build

```bash
npm run build    # type-checks with vue-tsc, then emits static assets to dist/
npm run preview
```

Set `VITE_GATEWAY_URL` to point the WHIP requests at a non-proxied gateway
origin in production (defaults to same-origin).

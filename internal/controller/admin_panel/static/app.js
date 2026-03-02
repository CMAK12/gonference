import { createPeer } from "./peer.js";
import { initLocalMedia } from "./media.js";
import { setupSignaling } from "./signaling.js";
import { createVideoElement } from "./video_manager.js";

const roomId = "room1";
const peerId = crypto.randomUUID();

const ws = new WebSocket("ws://localhost:8080/ws");

const peer = createPeer(ws, roomId, peerId, createVideoElement);

setupSignaling(ws, peer, roomId, peerId);

(async () => {
    await initLocalMedia(peer.pc, createVideoElement);

    console.log('creating offer with peerId:', peerId);

    const offer = await peer.pc.createOffer();
    await peer.pc.setLocalDescription(offer);

    console.log('sending offer to server');

    ws.send(JSON.stringify({
        type: "offer",
        roomId: roomId,
        memberId: peerId,
        sdp: offer.sdp
    }));

    console.log('negotiation complete');
})();
import { createPeer } from "./peer.js";
import { initLocalMedia } from "./media.js";
import {MessageType, setupSignaling} from "./signaling.js";
import { createVideoElement } from "./video_manager.js";

const roomId = "room1";
const peerId = crypto.randomUUID();

const ws = new WebSocket("ws://localhost:8080/ws");

const peer = createPeer(ws, roomId, peerId, createVideoElement);

setupSignaling(ws, peer, roomId, peerId);

(async () => {
    window.addEventListener("beforeunload", () => {
        try {
            ws.send(JSON.stringify({
                type: MessageType.LEAVE,
                roomId: roomId,
                memberId: peerId
            }));
        } catch (e) {}

        try {
            peer.pc.close()
        } catch (e) {}

        try {
            ws.close(1001, "Leaving room");
        } catch (e) {}
    })
    await initLocalMedia(peer.pc, createVideoElement);

    console.log('creating offer with peerId:', peerId);

    const offer = await peer.pc.createOffer();
    await peer.pc.setLocalDescription(offer);

    console.log('sending offer to server');

    ws.send(JSON.stringify({
        type: MessageType.OFFER,
        roomId: roomId,
        memberId: peerId,
        sdp: offer.sdp
    }));

    console.log('negotiation complete');
})();
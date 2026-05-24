import { createPeer } from "./peer.js";
import { initLocalMedia } from "./media.js";
import {MessageType, setupSignaling} from "./signaling.js";
import { createVideoElement } from "./video_manager.js";
import { initControls, showLeftOverlay } from "./controls.js";

const roomId = document.body.dataset.roomId;
const peerId = document.body.dataset.username;
const restPort = document.body.dataset.restPort;

const ws = new WebSocket(`ws://${location.hostname}:${restPort}/ws`);

const peer = createPeer(ws, roomId, peerId, createVideoElement);

setupSignaling(ws, peer, roomId, peerId);

let left = false;
let localTracks = { audioTrack: null, videoTrack: null };

function leave({ showOverlay = false } = {}) {
    if (left) return;
    left = true;

    try {
        ws.send(JSON.stringify({
            type: MessageType.LEAVE,
            roomId: roomId,
            memberId: peerId
        }));
    } catch (e) {}

    try {
        fetch(
            `http://${location.hostname}:${restPort}/conference/${encodeURIComponent(roomId)}/leave?username=${encodeURIComponent(peerId)}`,
            { method: "DELETE", keepalive: true }
        );
    } catch (e) {}

    try { peer.pc.close(); } catch (e) {}
    try { ws.close(1001, "Leaving room"); } catch (e) {}

    try { localTracks.audioTrack && localTracks.audioTrack.stop(); } catch (e) {}
    try { localTracks.videoTrack && localTracks.videoTrack.stop(); } catch (e) {}

    if (showOverlay) {
        const videos = document.getElementById('videos');
        while (videos.firstChild) videos.removeChild(videos.firstChild);
        document.querySelector('.controlbar').hidden = true;
        showLeftOverlay();
    }
}

window.addEventListener("beforeunload", () => leave());

(async () => {
    localTracks = await initLocalMedia(peer.pc, createVideoElement, peerId);

    initControls({
        audioTrack: localTracks.audioTrack,
        videoTrack: localTracks.videoTrack,
        onLeave: () => leave({ showOverlay: true }),
    });

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

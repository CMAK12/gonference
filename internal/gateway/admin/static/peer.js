import {MessageType} from "./signaling.js";

export function createPeer(ws, roomId, peerId, createVideoElement) {
    const pc = new RTCPeerConnection({
        iceServers: [{ urls: 'stun:stun.l.google.com:19302' }]
    });

    let pendingCandidates = [];

    pc.oniceconnectionstatechange = () => {
        console.log('ICE connection state:', pc.iceConnectionState);
    };

    pc.ontrack = (event) => {
        const stream = event.streams[0];

        console.log('ontrack event:', {
            trackKind: event.track.kind,
            trackId: event.track.id,
            streamId: stream?.id,
            myPeerId: peerId,
            isMyStream: stream?.id === peerId
        });

        if (!stream) return;

        console.log('creating video for remote stream:', stream.id);
        createVideoElement(stream);
    };

    pc.onicecandidate = (event) => {
        if (!event.candidate) return;

        ws.send(JSON.stringify({
            type: MessageType.CANDIDATE,
            roomId: roomId,
            memberId: peerId,
            candidate: event.candidate
        }));
    };

    pc.onnegotiationneeded = (event) => {
        console.log("negotiationneeded event:", event);
    };

    return {
        pc,
        getPendingCandidates: () => pendingCandidates,
        clearPending: () => { pendingCandidates = []; },
        pushPending: (c) => pendingCandidates.push(c)
    };
}
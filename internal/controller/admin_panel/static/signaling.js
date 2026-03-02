export const MessageType = Object.freeze({
    OFFER: "offer",
    ANSWER: "answer",
    CANDIDATE: "candidate",
    LEAVE: "leave"
});

export function setupSignaling(ws, peer, roomId, peerId) {
    ws.onmessage = async (event) => {
        console.log(event.data);
        const msg = JSON.parse(event.data);

        switch (msg.type) {
            case MessageType.OFFER:
                await peer.pc.setRemoteDescription({
                    type: MessageType.OFFER,
                    sdp: msg.sdp
                });

                const answer = await peer.pc.createAnswer();
                await peer.pc.setLocalDescription(answer);

                ws.send(JSON.stringify({
                    type: MessageType.ANSWER,
                    roomId: roomId,
                    memberId: peerId,
                    sdp: peer.pc.localDescription.sdp
                }));
                break;

            case MessageType.ANSWER:
                console.log('received answer from server');

                await peer.pc.setRemoteDescription({
                    type: MessageType.ANSWER,
                    sdp: msg.sdp
                });

                for (const c of peer.getPendingCandidates()) {
                    try {
                        await peer.pc.addIceCandidate(c);
                    } catch (e) {
                        console.warn('failed to add queued candidate', e);
                    }
                }

                peer.clearPending();
                break;

            case MessageType.CANDIDATE:
                const cand = msg.candidate;

                if (!peer.pc.remoteDescription) {
                    peer.pushPending(cand);
                    return;
                }

                try {
                    await peer.pc.addIceCandidate(cand);
                } catch (e) {
                    console.warn('failed to add candidate', e);
                }
                break;

            case MessageType.LEAVE:

        }
    };
}
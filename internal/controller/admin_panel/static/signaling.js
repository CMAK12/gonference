export function setupSignaling(ws, peer, roomId, peerId) {
    ws.onmessage = async (event) => {
        console.log(event.data);
        const msg = JSON.parse(event.data);

        switch (msg.type) {
            case "answer":
                console.log('received answer from server');

                await peer.pc.setRemoteDescription({
                    type: "answer",
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

            case "offer":
                await peer.pc.setRemoteDescription({
                    type: "offer",
                    sdp: msg.sdp
                });

                const answer = await peer.pc.createAnswer();
                await peer.pc.setLocalDescription(answer);

                ws.send(JSON.stringify({
                    type: "answer",
                    roomId: roomId,
                    memberId: peerId,
                    sdp: peer.pc.localDescription.sdp
                }));
                break;

            case "candidate":
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
        }
    };
}
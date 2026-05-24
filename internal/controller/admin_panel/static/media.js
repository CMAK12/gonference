export async function initLocalMedia(pc, createVideoElement, name) {
    const stream = await navigator.mediaDevices.getUserMedia({
        video: true,
        audio: true
    });

    createVideoElement(stream, { muted: true, name: name ? `${name} (you)` : 'you', local: true });

    stream.getTracks().forEach(track => {
        pc.addTrack(track, stream);
    });

    return {
        stream,
        audioTrack: stream.getAudioTracks()[0] || null,
        videoTrack: stream.getVideoTracks()[0] || null,
    };
}

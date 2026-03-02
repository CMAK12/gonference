export async function initLocalMedia(pc, createVideoElement) {
    const stream = await navigator.mediaDevices.getUserMedia({
        video: true,
        audio: false
    });

    createVideoElement(stream, true);

    stream.getTracks().forEach(track => {
        pc.addTrack(track, stream);
    });
}
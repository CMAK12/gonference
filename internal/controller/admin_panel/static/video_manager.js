const container = document.getElementById('videos');

export function updateLayout() {
    const count = container.children.length;

    if (count === 1) {
        container.style.gridTemplateColumns = '1fr';
        container.style.gridTemplateRows = '1fr';
        return;
    }

    if (count === 2) {
        container.style.gridTemplateColumns = '1fr 1fr';
        container.style.gridTemplateRows = '1fr';
        return;
    }

    if (count <= 4) {
        container.style.gridTemplateColumns = '1fr 1fr';
        container.style.gridTemplateRows = '1fr 1fr';
        return;
    }

    container.style.gridTemplateColumns =
        'repeat(auto-fill, minmax(300px, 1fr))';
    container.style.gridAutoRows = '1fr';
}

export function createVideoElement(stream, muted = false) {
    const video = document.createElement('video');
    video.id = `video-${stream.id}`;
    video.srcObject = stream;
    video.autoplay = true;
    video.playsInline = true;
    video.muted = muted;

    container.appendChild(video);
    updateLayout();
}

export function removeVideoElement(peerId) {
    const video = document.getElementById(`video-${peerId}`);
    if (!video) {
        console.warn("Video not found for peer: ", peerId);
        return;
    }

    video.srcObject = null;
    video.remove();
    updateLayout();
}
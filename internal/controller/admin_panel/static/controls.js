export function initControls({ audioTrack, videoTrack, onLeave }) {
    const micBtn = document.getElementById('btn-mic');
    const camBtn = document.getElementById('btn-cam');
    const leaveBtn = document.getElementById('btn-leave');

    function applyState(btn, enabled) {
        btn.classList.toggle('off', !enabled);
        btn.setAttribute('aria-pressed', String(!enabled));
    }

    if (audioTrack) {
        applyState(micBtn, audioTrack.enabled);
        micBtn.addEventListener('click', () => {
            audioTrack.enabled = !audioTrack.enabled;
            applyState(micBtn, audioTrack.enabled);
        });
    } else {
        micBtn.disabled = true;
    }

    if (videoTrack) {
        applyState(camBtn, videoTrack.enabled);
        camBtn.addEventListener('click', () => {
            videoTrack.enabled = !videoTrack.enabled;
            applyState(camBtn, videoTrack.enabled);
        });
    } else {
        camBtn.disabled = true;
    }

    leaveBtn.addEventListener('click', () => {
        onLeave();
    });
}

export function showLeftOverlay() {
    const overlay = document.getElementById('left-overlay');
    overlay.hidden = false;

    document.getElementById('btn-rejoin').addEventListener('click', () => {
        location.reload();
    });
    document.getElementById('btn-lobby').addEventListener('click', () => {
        location.href = '/';
    });
}

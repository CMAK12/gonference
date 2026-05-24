const container = document.getElementById('videos');

const GAP = 8;
const TARGET_AR = 16 / 9;

function bestLayout(n, w, h) {
    let best = null;
    for (let rows = 1; rows <= n; rows++) {
        const cols = Math.ceil(n / rows);
        const tileW = (w - GAP * (cols + 1)) / cols;
        const tileH = (h - GAP * (rows + 1)) / rows;
        if (tileW <= 0 || tileH <= 0) continue;

        const effW = Math.min(tileW, tileH * TARGET_AR);
        const effH = effW / TARGET_AR;
        const area = effW * effH;

        const empty = rows * cols - n;
        const score = area - (empty * area * 0.5);

        if (!best || score > best.score) {
            best = { rows, cols, score };
        }
    }
    return best || { rows: 1, cols: Math.max(1, n) };
}

export function updateLayout() {
    const count = container.children.length;
    if (count === 0) return;

    const rect = container.getBoundingClientRect();
    const { rows, cols } = bestLayout(count, rect.width, rect.height);

    container.style.gridTemplateColumns = `repeat(${cols}, 1fr)`;
    container.style.gridTemplateRows = `repeat(${rows}, 1fr)`;
    container.style.placeContent = 'center';
}

export function createVideoElement(stream, { muted = false, name = '', local = false } = {}) {
    const id = `tile-${stream.id}`;
    if (document.getElementById(id)) return;

    const tile = document.createElement('div');
    tile.id = id;
    tile.className = local ? 'tile local' : 'tile';

    const video = document.createElement('video');
    video.srcObject = stream;
    video.autoplay = true;
    video.playsInline = true;
    video.muted = muted;
    tile.appendChild(video);

    if (name) {
        const label = document.createElement('div');
        label.className = 'name';
        label.textContent = name.replace(/_/g, ' ');
        tile.appendChild(label);
    }

    container.appendChild(tile);
    updateLayout();
}

export function removeVideoElement(peerId) {
    const tile = document.getElementById(`tile-${peerId}`);
    if (!tile) {
        console.warn("Tile not found for peer: ", peerId);
        return;
    }

    const video = tile.querySelector('video');
    if (video) video.srcObject = null;
    tile.remove();
    updateLayout();
}

let resizeTimer = null;
window.addEventListener('resize', () => {
    if (resizeTimer) clearTimeout(resizeTimer);
    resizeTimer = setTimeout(updateLayout, 100);
});

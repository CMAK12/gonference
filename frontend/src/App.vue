<script setup lang="ts">
import { computed, onBeforeUnmount, ref } from 'vue'

import VideoGrid from './components/VideoGrid.vue'
import { useConference } from './composables/useConference'
import { useTheme } from './composables/useTheme'
import wordmarkDark from './assets/gonference-wordmark-dark.svg'
import wordmarkLight from './assets/gonference-wordmark-light.svg'

const room = ref('room1')

const { theme, toggleTheme } = useTheme()

// The dark wordmark has light text (for dark backgrounds); the light wordmark
// has dark text (for light backgrounds).
const wordmark = computed(() => (theme.value === 'dark' ? wordmarkDark : wordmarkLight))

const { connected, connecting, error, memberId, localStream, remotePeers, join, leave } =
  useConference()

function onJoin(): void {
  const name = room.value.trim()
  if (name) join(name)
}

function onBeforeUnload(): void {
  leave()
}

window.addEventListener('beforeunload', onBeforeUnload)
onBeforeUnmount(() => window.removeEventListener('beforeunload', onBeforeUnload))
</script>

<template>
  <div class="app">
    <header class="bar">
      <h1 class="bar__title">
        <img :src="wordmark" alt="gonference" class="bar__logo" />
      </h1>

      <button
        class="btn btn--icon"
        type="button"
        :aria-label="theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'"
        :title="theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'"
        @click="toggleTheme"
      >
        {{ theme === 'dark' ? '☀️' : '🌙' }}
      </button>

      <div v-if="!connected" class="join">
        <input
          v-model="room"
          class="join__input"
          :disabled="connecting"
          placeholder="Room name"
          @keyup.enter="onJoin"
        />
        <button class="btn" :disabled="connecting || !room.trim()" @click="onJoin">
          {{ connecting ? 'Joining…' : 'Join' }}
        </button>
      </div>

      <div v-else class="session">
        <span class="session__meta">Room {{ room }} · you {{ memberId.slice(0, 8) }}</span>
        <button class="btn btn--danger" @click="leave">Leave</button>
      </div>
    </header>

    <p v-if="error" class="error">{{ error }}</p>

    <main class="main">
      <VideoGrid
        v-if="connected"
        :local-stream="localStream"
        :remote-peers="remotePeers"
      />
      <div v-else class="placeholder">
        <p>Enter a room name and join to start streaming.</p>
      </div>
    </main>
  </div>
</template>

<style scoped>
.app {
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
}

.bar {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 16px;
}

.bar__title {
  margin: 0 auto 0 0;
  line-height: 0;
}

.bar__logo {
  display: block;
  height: 52px;
  width: auto;
}

.join,
.session {
  display: flex;
  align-items: center;
  gap: 8px;
}

.session__meta {
  font-size: 13px;
  color: var(--text-muted);
}

.join__input {
  padding: 8px 10px;
  border: 1px solid var(--border);
  border-radius: 8px;
  background: var(--surface);
  color: var(--text);
}

.join__input::placeholder {
  color: var(--placeholder);
}

.btn {
  padding: 8px 16px;
  border: 0;
  border-radius: 8px;
  background: var(--primary);
  color: var(--primary-text);
  font-weight: 600;
  cursor: pointer;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn--danger {
  background: var(--danger);
  color: var(--danger-text);
}

.btn--icon {
  padding: 8px;
  width: 38px;
  background: var(--surface);
  border: 1px solid var(--border);
  font-size: 16px;
  line-height: 1;
}

.error {
  margin: 0 0 16px;
  padding: 10px 12px;
  border-radius: 8px;
  background: var(--error-bg);
  color: var(--error-text);
  font-size: 14px;
}

.placeholder {
  display: grid;
  place-items: center;
  min-height: 40vh;
  color: var(--placeholder);
}
</style>

<script setup lang="ts">
import { computed } from 'vue'

import VideoTile from './VideoTile.vue'
import type { RemotePeer } from '../composables/useConference'

const props = defineProps<{
  localStream: MediaStream | null
  remotePeers: Record<string, RemotePeer>
}>()

const remotes = computed(() => Object.values(props.remotePeers))
const tileCount = computed(() => remotes.value.length + (props.localStream ? 1 : 0))
</script>

<template>
  <div class="grid" :data-count="tileCount">
    <VideoTile v-if="localStream" :stream="localStream" :muted="true" label="You" />
    <VideoTile
      v-for="peer in remotes"
      :key="peer.memberId"
      :stream="peer.stream"
      :label="peer.memberId.slice(0, 8)"
    />
  </div>
</template>

<style scoped>
.grid {
  display: grid;
  gap: 12px;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  align-content: start;
}

.grid[data-count='1'] {
  grid-template-columns: minmax(0, 720px);
}

.grid[data-count='2'] {
  grid-template-columns: 1fr 1fr;
}
</style>

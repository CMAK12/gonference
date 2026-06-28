<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'

const props = defineProps<{
  stream: MediaStream
  label?: string
  muted?: boolean
}>()

const video = ref<HTMLVideoElement | null>(null)

function bind(): void {
  if (!video.value) return
  // srcObject must be assigned as a property, not a template attribute.
  video.value.srcObject = props.stream
  video.value.muted = props.muted ?? false
}

onMounted(bind)
watch(() => props.stream, bind)
</script>

<template>
  <div class="tile">
    <video ref="video" autoplay playsinline></video>
    <span v-if="label" class="tile__label">{{ label }}</span>
  </div>
</template>

<style scoped>
.tile {
  position: relative;
  aspect-ratio: 16 / 9;
  background: var(--tile-bg);
  border-radius: 10px;
  overflow: hidden;
}

.tile video {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.tile__label {
  position: absolute;
  left: 8px;
  bottom: 8px;
  padding: 2px 8px;
  font-size: 12px;
  border-radius: 6px;
  background: var(--label-bg);
  color: var(--label-text);
}
</style>

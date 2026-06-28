import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// The gateway exposes the WHIP endpoint on :8080. In development we proxy
// /whip there so the SPA and the API share an origin (no CORS needed).
export default defineConfig({
  plugins: [vue()],
  server: {
    port: 5173,
    proxy: {
      '/whip': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
})

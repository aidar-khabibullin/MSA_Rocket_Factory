import { fileURLToPath } from 'node:url'

import tailwindcss from '@tailwindcss/vite'
import react from '@vitejs/plugin-react'
import { defineConfig } from 'vitest/config'

// Dev-прокси решает CORS: браузер ходит только на origin Vite, а все /api уходят на единую точку входа, Envoy.
// Envoy сам маршрутизирует: /api/auth, /api/parts -> iam и inventory (gRPC), /api/v1 -> order (HTTP).
export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: { '@': fileURLToPath(new URL('./src', import.meta.url)) },
  },
  server: {
    proxy: {
      '/api': process.env.API_PROXY_TARGET ?? 'http://localhost:8085',
    },
  },
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: ['./src/global/test-setup.ts'],
    css: false,
  },
})

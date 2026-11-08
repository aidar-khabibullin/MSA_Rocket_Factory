import { fileURLToPath } from 'node:url'

import tailwindcss from '@tailwindcss/vite'
import react from '@vitejs/plugin-react'
import { defineConfig } from 'vitest/config'

// Dev-прокси решает CORS: браузер ходит только на origin Vite.
//  /api/auth, /api/parts, /api/health -> bff (gRPC-шлюз к iam и inventory)
//  /api/v1                            -> order (HTTP)
export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: { '@': fileURLToPath(new URL('./src', import.meta.url)) },
  },
  server: {
    proxy: {
      '/api/auth': 'http://localhost:8095',
      '/api/parts': 'http://localhost:8095',
      '/api/health': 'http://localhost:8095',
      '/api/v1': 'http://localhost:8080',
    },
  },
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: ['./src/global/test-setup.ts'],
    css: false,
  },
})

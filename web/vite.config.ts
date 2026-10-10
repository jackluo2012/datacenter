import { fileURLToPath, URL } from 'node:url'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  server: {
    host: true,
    port: 5173,
    // 网关(datacenter-api, :8887)已开 CORS，前端直连后端，不设 /api 代理；
    // 后端地址在 src/api/http.ts 的 baseURL（可用 VITE_API_BASE 覆盖）
  },
  build: {
    chunkSizeWarningLimit: 1500,
  },
})

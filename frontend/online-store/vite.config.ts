import { fileURLToPath, URL } from 'node:url'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-vue-components/vite'
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers'

const apiGateway = process.env.VITE_API_GATEWAY || 'http://127.0.0.1:18000'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    AutoImport({
      imports: ['vue', 'vue-router', 'pinia'],
      resolvers: [ElementPlusResolver()],
      dts: 'src/auto-imports.d.ts',
    }),
    Components({
      resolvers: [ElementPlusResolver()],
      dts: 'src/components.d.ts',
    }),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  server: {
    port: 4173,
    proxy: {
      // user_web routes
      '/u': {
        target: apiGateway,
        changeOrigin: true,
      },
      // goods_web routes
      '/g': {
        target: apiGateway,
        changeOrigin: true,
      },
      // order_web routes
      '/o': {
        target: apiGateway,
        changeOrigin: true,
      },
      // userop_web routes
      '/uo': {
        target: apiGateway,
        changeOrigin: true,
      },
    },
  },
})

import { fileURLToPath, URL } from 'node:url'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-vue-components/vite'
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers'

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
      // user_web (local :8021) — auth, register, captcha, sms, user
      '/u': {
        target: 'http://127.0.0.1:8021',
        changeOrigin: true,
      },
      // goods_web (local :8022) — goods, category, banner, brand
      '/g': {
        target: 'http://127.0.0.1:8022',
        changeOrigin: true,
      },
      // order_web (local :8023) — cart, orders, alipay
      '/o': {
        target: 'http://127.0.0.1:8023',
        changeOrigin: true,
      },
      // userop_web (local :8027) — favs, addresses, messages
      '/uo': {
        target: 'http://127.0.0.1:8027',
        changeOrigin: true,
      },
      // external legacy host — index goods grouping & hot search keywords
      '/ext': {
        target: 'http://shop.projectsedu.com',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/ext/, ''),
      },
    },
  },
})

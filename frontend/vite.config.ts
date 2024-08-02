import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react-swc';
import path from 'path';
import tsconfigPaths from 'vite-tsconfig-paths';
import { VitePWA } from 'vite-plugin-pwa';

export default defineConfig({
  plugins: [
    react(),
    tsconfigPaths(),
    VitePWA({
      registerType: 'autoUpdate',
      includeAssets: ['favicon.ico', 'robots.txt', 'apple-touch-icon.png'],
      manifest: {
        short_name: 'Alby Hub',
        name: 'Alby Hub',
        icons: [
          {
            src: 'favicon.svg',
            sizes: 'any',
            type: 'image/svg+xml'
          }
        ],
        start_url: '.',
        display: 'standalone',
        theme_color: '#000000',
        background_color: '#ffffff'
      },
      workbox: {
        globPatterns: ['**/*.{js,css,html,png,svg,ico}']
      }
    })
  ],
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        secure: false,
      },
    },
  },
  resolve: {
    alias: {
      src: path.resolve(__dirname, './src'),
      wailsjs: path.resolve(__dirname, './wailsjs'),
    },
  },
});

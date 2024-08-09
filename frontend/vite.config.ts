import react from "@vitejs/plugin-react-swc";
import path from "path";
import { defineConfig, Plugin } from "vite";
import { VitePWA } from "vite-plugin-pwa";
import tsconfigPaths from "vite-tsconfig-paths";

export default defineConfig(({ command }) => ({
  plugins: [
    react(),
    tsconfigPaths(),
    VitePWA({
      registerType: "autoUpdate",
      includeAssets: [
        "favicon.ico",
        "robots.txt",
        "icon-512.png",
        "icon-192.png",
      ],
      manifest: {
        short_name: "Alby Hub",
        name: "Alby Hub",
        icons: [
          {
            src: "favicon.svg",
            sizes: "any",
            type: "image/svg+xml",
          },
          {
            src: "icon-192.png",
            sizes: "192x192",
            type: "image/png",
          },
          {
            src: "icon-512.png",
            sizes: "512x512",
            type: "image/png",
          },
        ],
        start_url: ".",
        display: "standalone",
        theme_color: "#000000",
        background_color: "#ffffff",
      },
      workbox: {
        globPatterns: ["**/*.{js,css,html,png,svg,ico}"],
      },
    }),
    ...(command === "serve" ? [insertDevCSPNoncePlugin] : []),
  ],
  server: {
    proxy: {
      "/api": {
        target: "http://localhost:8080",
        secure: false,
      },
    },
  },
  resolve: {
    alias: {
      src: path.resolve(__dirname, "./src"),
      wailsjs: path.resolve(__dirname, "./wailsjs"),
    },
  },
  html:
    command === "serve"
      ? {
          cspNonce: "PLACEHOLDER",
        }
      : undefined,
}));

const insertDevCSPNoncePlugin: Plugin = {
  name: "transform-html",
  transformIndexHtml: {
    enforce: "pre",
    transform(html) {
      return html.replace(
        "default-src 'self'",
        "default-src 'self' 'nonce-PLACEHOLDER'"
      );
    },
  },
};

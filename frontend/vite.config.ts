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
      // disable service worker - Alby Hub cannot be used offline (and also breaks oauth callback)
      injectRegister: false,
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
    }),
    ...(command === "serve" ? [insertDevCSPPlugin] : []),
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
  build: {
    assetsInlineLimit: 0,
  },
  html:
    command === "serve"
      ? {
          cspNonce: "DEVELOPMENT",
        }
      : undefined,
}));

const DEVELOPMENT_NONCE = "'nonce-DEVELOPMENT'";

const insertDevCSPPlugin: Plugin = {
  name: "dev-csp",
  transformIndexHtml: {
    enforce: "pre",
    transform(html) {
      return html.replace(
        "<head>",
        `<head>
        <!-- DEV-ONLY CSP - when making changes here, also update the CSP header in http_service.go (without the nonce!) -->
        <meta http-equiv="Content-Security-Policy" content="default-src 'self' ${DEVELOPMENT_NONCE}; img-src 'self' https://uploads.getalby-assets.com https://getalby.com; connect-src 'self' https://api.getalby.com"/>`
      );
    },
  },
};

import react from "@vitejs/plugin-react-swc";
import path from "path";
import { defineConfig } from "vite";
import tsconfigPaths from "vite-tsconfig-paths";

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react(), tsconfigPaths()],
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
});

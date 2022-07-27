import { defineConfig } from "vite";

import eslint from "vite-plugin-eslint";
import vue from "@vitejs/plugin-vue2";

import path from "path";

export default defineConfig({
  plugins: [vue(), eslint({ fix: true })],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
  server: {
    port: 33159,
    strictPort: true,
    proxy: {
      "/api": {
        target: "http://localhost:33160",
        changeOrigin: true,
        secure: false,
      },
      "/api-background": {
        target: "http://localhost:33160",
        changeOrigin: true,
        secure: false,
      },
      "/ssr": {
        target: "http://localhost:33160",
        changeOrigin: true,
        secure: false,
      },
    },
  },
});

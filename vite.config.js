import { defineConfig } from "vite";

import eslint from "vite-plugin-eslint";
import vue from "@vitejs/plugin-vue2";

import path from "path";

export default defineConfig({
  define: {
    "process.env.NODE_ENV": JSON.stringify(process.env.NODE_ENV),
    "process.env.BASE_URL": JSON.stringify(process.env.BASE_URL),
  },
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

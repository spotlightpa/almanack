import { defineConfig } from "vite";

import Components from "unplugin-vue-components/vite";
import vue from "@vitejs/plugin-vue";
import eslint from "vite-plugin-eslint";

import path from "path";

export default defineConfig({
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
  plugins: [
    Components({}),
    vue({
      template: {
        compilerOptions: {
          isCustomElement: (tag) => tag === "raw-html",
        },
      },
    }),
    eslint({ fix: true }),
  ],
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

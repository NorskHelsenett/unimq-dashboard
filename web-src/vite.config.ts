import { resolve } from "path";
import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      "@": resolve(__dirname, "./src"),
    },
  },
  build: {
    outDir: "../web-src/static/dist",
    emptyOutDir: true,
    rollupOptions: {
      input: {
        index: resolve(__dirname, "index.html"),
        queue: resolve(__dirname, "src/pages/Queue.tsx"),
        notifications: resolve(__dirname, "notifications.html"),
        notifRule: resolve(__dirname, "notification_rule.html"),
        maintenance: resolve(__dirname, "maintenance.html"),
        editMaintenance: resolve(__dirname, "edit_maintenance.html"),
        profile: resolve(__dirname, "profile.html"),
        callback: resolve(__dirname, "callback.html"),
      },
      output: {
        entryFileNames: "[name].js",
        chunkFileNames: "[name].js",
        assetFileNames: "[name].[ext]",
      },
    },
  },
  server: {
    proxy: {
      "/api": "http://localhost:8080",
      //TODO: Should ideally have a better way to serve the static html files for the notifications rules page
      "/notifications/rule": {
        target: "http://localhost:8080",
        bypass: (req) => {
          if (req.method === "GET") return "/notification_rule.html"
          return null
        },
      },
      "/notifications/": "http://localhost:8080",
      "/maintenance": {
        target: "http://localhost:8080",
        bypass: (req) => {
          if (req.method === "GET") {
            if (req.url?.startsWith("/maintenance/edit")) return "/edit_maintenance.html"
            return "/maintenance.html"
          }
          return null
        },
      },
      "/static/logo": "http://localhost:8080",
    },
  },
});

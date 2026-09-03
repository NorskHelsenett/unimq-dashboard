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
        notifications: resolve(__dirname, "entries/notifications.html"),
        notifRule: resolve(__dirname, "entries/notification_rule.html"),
        maintenance: resolve(__dirname, "entries/maintenance.html"),
        editMaintenance: resolve(__dirname, "entries/edit_maintenance.html"),
        profile: resolve(__dirname, "entries/profile.html"),
        callback: resolve(__dirname, "entries/callback.html"),
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
          if (req.method === "GET") return "/entries/notification_rule.html"
          return null
        },
      },
      "/notifications": {
        target: "http://localhost:8080",
        bypass: (req) => {
          if (req.method === "GET" && !req.url?.startsWith("/notifications/rule"))
            return "/entries/notifications.html"
          return null
        },
      },
      "/maintenance": {
        target: "http://localhost:8080",
        bypass: (req) => {
          if (req.method === "GET") {
            if (req.url?.startsWith("/maintenance/edit")) return "/entries/edit_maintenance.html"
            return "/entries/maintenance.html"
          }
          return null
        },
      },
      "/profile": {
        target: "http://localhost:8080",
        bypass: (req) => {
          if (req.method === "GET") return "/entries/profile.html"
          return null
        },
      },
    },
  },
});

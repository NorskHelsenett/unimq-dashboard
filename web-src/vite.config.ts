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
    outDir: "../web/static/dist",
    emptyOutDir: true,
    rollupOptions: {
      input: {
        index: resolve(__dirname, "src/pages/index.tsx"),
        queue: resolve(__dirname, "src/pages/Queue.tsx"),
        notifications: resolve(__dirname, "src/pages/Notifications.tsx"),
        notifRule: resolve(__dirname, "src/pages/NotifyRule.tsx"),
        maintenance: resolve(__dirname, "src/pages/Maintenance.tsx"),
        maintAdmin: resolve(__dirname, "src/pages/MaintenanceAdmin.tsx"),
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
      //Litt usikker på om dette går og er final ?
      "/notifications/rule": {
        target: "http://localhost:8080",
        bypass: (req) => {
          if (req.method === "GET") return "/notification_rule.html"
          return null
        },
      },
      "/notifications": "http://localhost:8080",
      "/notifications/": "http://localhost:8080",
      "/maintenance": "http://localhost:8080",
      "/static/logo": "http://localhost:8080",
    },
  },
});

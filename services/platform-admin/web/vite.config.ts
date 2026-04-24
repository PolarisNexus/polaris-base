import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

// 开发时前端 :5173，BFF :8080，APISIX :9080，通过代理避免 CORS。
// 生产走 APISIX 同源反代，/polaris.* 路由到 platform-admin。
export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      "/polaris.platform_admin.v1.": "http://localhost:8080",
      // AI Gateway 冒烟测试通道：页面直接打 /ai/v1/* 由 Vite 代理到 APISIX
      "/ai/v1/": "http://localhost:9080",
    },
  },
});

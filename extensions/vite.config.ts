import { defineConfig } from "vite";
import webExtension from "vite-plugin-web-extension";

export default defineConfig(({ mode }) => ({
  plugins: [
    webExtension({
      browser: mode === "firefox" ? "firefox" : "chrome",
    }),
  ],
}));

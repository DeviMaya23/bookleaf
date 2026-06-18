import { defineConfig } from "vite";
import webExtension from "vite-plugin-web-extension";

export default defineConfig(({ mode }) => {
  const isFirefox = mode === "firefox" || mode === "firefox-production";
  const isProduction = mode === "chrome-production" || mode === "firefox-production";
  const geckoId =
    mode === "firefox-production" ? "bookleaf@evimay.me" : "bookleaf-dev@evimay.me";
  return {
    build: {
      outDir: isFirefox ? "dist/firefox" : "dist/chrome",
    },
    plugins: [
      webExtension({
        browser: isFirefox ? "firefox" : "chrome",
        transformManifest: (manifest) => {
          const typed = manifest as typeof manifest & {
            name: string;
            background: { service_worker: string; type?: string };
          };
          return {
            ...typed,
            name: isProduction ? typed.name : `${typed.name} (Dev)`,
            ...(isFirefox && {
              background: {
                scripts: [typed.background.service_worker],
              },
              browser_specific_settings: {
                gecko: { id: geckoId },
              },
            }),
          };
        },
      }),
    ],
  };
});

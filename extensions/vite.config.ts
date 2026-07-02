import { defineConfig } from "vite";
import webExtension from "vite-plugin-web-extension";

const FIREFOX_UPDATE_URL = "https://bookleaf-files.evimay.me/bookleaf-extension-updates.json";

export default defineConfig(({ mode }) => {
  const isFirefox = mode === "firefox" || mode === "firefox-production";
  const isFirefoxProduction = mode === "firefox-production";
  const isProduction = mode === "chrome-production" || mode === "firefox-production";
  const geckoId = isFirefoxProduction ? "bookleaf@evimay.me" : "bookleaf-dev@evimay.me";
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
                gecko: {
                  id: geckoId,
                  ...(isFirefoxProduction && { update_url: FIREFOX_UPDATE_URL }),
                },
              },
            }),
          };
        },
      }),
    ],
  };
});

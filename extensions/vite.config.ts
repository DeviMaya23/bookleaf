import { defineConfig } from "vite";
import webExtension from "vite-plugin-web-extension";

export default defineConfig(({ mode }) => {
  const isFirefox = mode === "firefox";
  return {
    build: {
      outDir: isFirefox ? "dist/firefox" : "dist/chrome",
    },
    plugins: [
      webExtension({
        browser: isFirefox ? "firefox" : "chrome",
        ...(isFirefox && {
          transformManifest: (manifest) => {
            const { background, ...rest } = manifest as typeof manifest & {
              background: { service_worker: string; type?: string };
            };
            return {
              ...rest,
              background: {
                scripts: [background.service_worker],
              },
              browser_specific_settings: {
                gecko: { id: "bookleaf@evimay.me" },
              },
            };
          },
        }),
      }),
    ],
  };
});

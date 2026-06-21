import { useEffect, useState } from "react";
import browser from "webextension-polyfill";
import { Moon, Sun } from "lucide-react";
import { getDragEnabled, setDragEnabled } from "../lib/storage";
import type { Colors } from "./App";

const SNIP_COMMAND_NAME = "snip-capture";
const CHROME_SHORTCUTS_URL = "chrome://extensions/shortcuts";

export const isFirefox =
  import.meta.env.MODE === "firefox" || import.meta.env.MODE === "firefox-production";

export async function openFirefoxShortcutSettings(): Promise<void> {
  await browser.commands.openShortcutSettings();
}

export async function openChromeShortcutSettings(): Promise<void> {
  await browser.tabs.create({ url: CHROME_SHORTCUTS_URL });
}

export default function Settings({
  c,
  isDark,
  onToggleDark,
  onBack,
}: {
  c: Colors;
  isDark: boolean;
  onToggleDark: () => void;
  onBack: () => void;
}) {
  const [dragEnabled, setDragEnabledState] = useState(true);
  const [shortcut, setShortcut] = useState("");

  useEffect(() => {
    getDragEnabled().then(setDragEnabledState);
    browser.commands.getAll().then((commands) => {
      const snip = commands.find((command) => command.name === SNIP_COMMAND_NAME);
      setShortcut(snip?.shortcut ?? "");
    });
  }, []);

  async function handleToggleDrag() {
    const next = !dragEnabled;
    setDragEnabledState(next);
    await setDragEnabled(next);
  }

  async function handleChangeClick() {
    if (isFirefox) {
      await openFirefoxShortcutSettings();
      return;
    }
    await openChromeShortcutSettings();
  }

  return (
    <div
      style={{
        width: 320,
        background: c.bg,
        border: `1px solid ${c.border}`,
        fontFamily:
          '-apple-system, BlinkMacSystemFont, "Segoe UI", "Helvetica Neue", sans-serif',
        overflow: "hidden",
      }}
    >
      {/* Header */}
      <div
        style={{
          height: 44,
          display: "flex",
          alignItems: "center",
          padding: "0 14px",
          borderBottom: `1px solid ${c.border}`,
          gap: 9,
        }}
      >
        <button
          onClick={onBack}
          title="Back"
          style={{
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            width: 22,
            height: 22,
            color: c.textSec,
            background: "transparent",
            border: "none",
            cursor: "pointer",
            padding: 0,
            fontSize: 15,
          }}
        >
          ←
        </button>
        <span
          style={{
            fontSize: 14,
            fontWeight: 600,
            color: c.text,
            letterSpacing: "-0.02em",
          }}
        >
          Settings
        </span>
      </div>

      {/* Dark mode row */}
      <div
        style={{
          height: 48,
          display: "flex",
          alignItems: "center",
          padding: "0 14px",
          borderBottom: `1px solid ${c.divider}`,
          gap: 10,
        }}
      >
        <span style={{ fontSize: 13, color: c.text, flex: 1 }}>Dark mode</span>
        <button
          onClick={onToggleDark}
          title={isDark ? "Switch to light mode" : "Switch to dark mode"}
          style={{
            width: 28,
            height: 28,
            borderRadius: 6,
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            border: "none",
            cursor: "pointer",
            background: "transparent",
            color: c.textSec,
          }}
        >
          {isDark ? <Sun size={14} /> : <Moon size={14} />}
        </button>
      </div>

      {/* Drag-to-save row */}
      <div
        style={{
          height: 48,
          display: "flex",
          alignItems: "center",
          padding: "0 14px",
          borderBottom: `1px solid ${c.divider}`,
          gap: 10,
        }}
      >
        <span style={{ fontSize: 13, color: c.text, flex: 1 }}>Drag to save</span>
        <button
          onClick={handleToggleDrag}
          style={{
            fontSize: 12,
            color: dragEnabled ? c.accent : c.textSec,
            background: "transparent",
            border: `1px solid ${c.border}`,
            borderRadius: 6,
            padding: "4px 9px",
            cursor: "pointer",
          }}
        >
          {dragEnabled ? "On" : "Off"}
        </button>
      </div>

      {/* Snip hotkey row */}
      <div
        style={{
          height: 48,
          display: "flex",
          alignItems: "center",
          padding: "0 14px",
          gap: 10,
        }}
      >
        <span style={{ fontSize: 13, color: c.text, flex: 1 }}>Snip hotkey</span>
        <button
          onClick={handleChangeClick}
          style={{
            fontSize: 12,
            color: c.textSec,
            background: "transparent",
            border: `1px solid ${c.border}`,
            borderRadius: 6,
            padding: "4px 9px",
            cursor: "pointer",
          }}
        >
          {shortcut || "Not set"}
        </button>
      </div>
    </div>
  );
}

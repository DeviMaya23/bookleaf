import { useEffect, useState } from "react";
import browser from "webextension-polyfill";
import { getDragEnabled, setDragEnabled, getDefaultFolder, setDefaultFolder, clearDefaultFolder } from "../lib/storage";
import { getFolders } from "../lib/api";
import type { Colors } from "./App";

const SNIP_COMMAND_NAME = "snip-capture";
const BROWSE_IMAGES_COMMAND_NAME = "browse-images";
const CHROME_SHORTCUTS_URL = "chrome://extensions/shortcuts";

export const isFirefox =
  import.meta.env.MODE === "firefox" || import.meta.env.MODE === "firefox-production";

export async function openFirefoxShortcutSettings(): Promise<void> {
  await browser.commands.openShortcutSettings();
}

export async function openChromeShortcutSettings(): Promise<void> {
  await browser.tabs.create({ url: CHROME_SHORTCUTS_URL });
}

function Switch({
  c,
  checked,
  onChange,
  title,
}: {
  c: Colors;
  checked: boolean;
  onChange: () => void;
  title?: string;
}) {
  return (
    <button
      role="switch"
      aria-checked={checked}
      onClick={onChange}
      title={title}
      style={{
        width: 32,
        height: 20,
        borderRadius: 10,
        border: "none",
        padding: 2,
        cursor: "pointer",
        background: checked ? c.accent : c.border,
        display: "flex",
        alignItems: "center",
        justifyContent: checked ? "flex-end" : "flex-start",
        transition: "background 0.15s",
      }}
    >
      <span
        style={{
          width: 16,
          height: 16,
          borderRadius: "50%",
          background: c.bg,
          transition: "transform 0.15s",
        }}
      />
    </button>
  );
}

export default function Settings({
  c,
  isDark,
  onToggleDark,
  onBack,
  onLogout,
}: {
  c: Colors;
  isDark: boolean;
  onToggleDark: () => void;
  onBack: () => void;
  onLogout: () => void;
}) {
  const [dragEnabled, setDragEnabledState] = useState(true);
  const [shortcut, setShortcut] = useState("");
  const [browseShortcut, setBrowseShortcut] = useState("");
  const [folders, setFolders] = useState<Array<{ id: string; name: string }>>([]);
  const [foldersLoading, setFoldersLoading] = useState(true);
  const [defaultFolderId, setDefaultFolderIdState] = useState("");

  useEffect(() => {
    getDragEnabled().then(setDragEnabledState);
    browser.commands.getAll().then((commands) => {
      const snip = commands.find((command) => command.name === SNIP_COMMAND_NAME);
      setShortcut(snip?.shortcut ?? "");
      const browse = commands.find((command) => command.name === BROWSE_IMAGES_COMMAND_NAME);
      setBrowseShortcut(browse?.shortcut ?? "");
    });
    getDefaultFolder().then((folder) => {
      setDefaultFolderIdState(folder?.id ?? "");
    });
    getFolders()
      .then(setFolders)
      .catch(() => {
        // leave folders empty; stored preference is preserved
      })
      .finally(() => setFoldersLoading(false));
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

  async function handleFolderChange(e: React.ChangeEvent<HTMLSelectElement>) {
    const value = e.target.value;
    if (value === "") {
      setDefaultFolderIdState("");
      await clearDefaultFolder();
    } else {
      const folder = folders.find((f) => f.id === value);
      if (!folder) return;
      setDefaultFolderIdState(value);
      await setDefaultFolder(folder);
    }
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
          gap: 10,
        }}
      >
        <span style={{ fontSize: 13, color: c.text, flex: 1 }}>Dark mode</span>
        <Switch
          c={c}
          checked={isDark}
          onChange={onToggleDark}
          title={isDark ? "Switch to light mode" : "Switch to dark mode"}
        />
      </div>

      {/* Drag-to-save row */}
      <div
        style={{
          height: 48,
          display: "flex",
          alignItems: "center",
          padding: "0 14px",
          gap: 10,
        }}
      >
        <span style={{ fontSize: 13, color: c.text, flex: 1 }}>Drag to save</span>
        <Switch c={c} checked={dragEnabled} onChange={handleToggleDrag} />
      </div>

      {/* Save destination section header */}
      <div
        style={{
          padding: "10px 14px 4px",
          borderTop: `1px solid ${c.border}`,
        }}
      >
        <span
          style={{
            fontSize: 10,
            fontWeight: 600,
            letterSpacing: "0.07em",
            textTransform: "uppercase",
            color: c.textSec,
          }}
        >
          Save destination
        </span>
      </div>

      {/* Default folder row */}
      <div
        style={{
          height: 48,
          display: "flex",
          alignItems: "center",
          padding: "0 14px",
          gap: 10,
        }}
      >
        <span style={{ fontSize: 13, color: c.text, flex: 1 }}>Default folder</span>
        <select
          disabled={foldersLoading}
          value={defaultFolderId}
          onChange={handleFolderChange}
          style={{
            fontSize: 12,
            color: c.text,
            background: c.bg,
            border: `1px solid ${c.border}`,
            borderRadius: 6,
            padding: "4px 9px",
            cursor: foldersLoading ? "default" : "pointer",
            maxWidth: 140,
          }}
        >
          <option value="">None (Unsorted)</option>
          {folders.map((folder) => (
            <option key={folder.id} value={folder.id}>
              {folder.name}
            </option>
          ))}
        </select>
      </div>

      {/* Hotkeys section header */}
      <div
        style={{
          padding: "10px 14px 4px",
          borderTop: `1px solid ${c.border}`,
        }}
      >
        <span
          style={{
            fontSize: 10,
            fontWeight: 600,
            letterSpacing: "0.07em",
            textTransform: "uppercase",
            color: c.textSec,
          }}
        >
          Hotkeys
        </span>
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
        <span style={{ fontSize: 13, color: c.text, flex: 1 }}>Snip</span>
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

      {/* Batch Save hotkey row */}
      <div
        style={{
          height: 48,
          display: "flex",
          alignItems: "center",
          padding: "0 14px",
          gap: 10,
        }}
      >
        <span style={{ fontSize: 13, color: c.text, flex: 1 }}>Batch Save</span>
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
          {browseShortcut || "Not set"}
        </button>
      </div>

      {/* Log out row */}
      <div
        style={{
          height: 48,
          display: "flex",
          alignItems: "center",
          justifyContent: "flex-end",
          padding: "0 14px",
          borderTop: `1px solid ${c.border}`,
        }}
      >
        <button
          onClick={onLogout}
          style={{
            fontSize: 12,
            color: "#e53e3e",
            background: "transparent",
            border: "none",
            cursor: "pointer",
            padding: 0,
            letterSpacing: "-0.01em",
          }}
        >
          Log out
        </button>
      </div>
    </div>
  );
}

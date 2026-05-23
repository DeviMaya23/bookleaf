import { useEffect, useState } from "react";
import browser from "webextension-polyfill";
import { login } from "../lib/auth";
import {
  clearAuth,
  getAuth,
  getAvatar,
  getDarkMode,
  getRecentSaves,
  getUsername,
  setDarkMode,
  type RecentSave,
} from "../lib/storage";

type AuthState = "loading" | "unauthenticated" | "authenticated";

interface Colors {
  bg: string;
  border: string;
  divider: string;
  text: string;
  textSec: string;
  textTer: string;
  hover: string;
  accent: string;
  thumbBorder: string;
}

const light: Colors = {
  bg: "#fff",
  border: "#e8e8e8",
  divider: "#f0f0f0",
  text: "#1c1c1c",
  textSec: "#aaa",
  textTer: "#ddd",
  hover: "#f5f5f5",
  accent: "#1c1c1c",
  thumbBorder: "rgba(0,0,0,0.06)",
};

const dark: Colors = {
  bg: "#1c1c1c",
  border: "#303030",
  divider: "#272727",
  text: "#efefef",
  textSec: "#6e6e6e",
  textTer: "#3a3a3a",
  hover: "#282828",
  accent: "#efefef",
  thumbBorder: "rgba(255,255,255,0.07)",
};

function MoonIcon() {
  return (
    <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
      <path
        d="M11.8 8.4A5.2 5.2 0 0 1 5.6 2.2a5 5 0 1 0 6.2 6.2Z"
        fill="currentColor"
      />
    </svg>
  );
}

function SunIcon() {
  return (
    <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
      <circle cx="7" cy="7" r="2.6" fill="currentColor" />
      <path
        d="M7 1.2V2.8M7 11.2V12.8M1.2 7H2.8M11.2 7H12.8M3.1 3.1L4.2 4.2M9.8 9.8L10.9 10.9M10.9 3.1L9.8 4.2M4.2 9.8L3.1 10.9"
        stroke="currentColor"
        strokeWidth="1.25"
        strokeLinecap="round"
      />
    </svg>
  );
}

export default function App() {
  const [authState, setAuthState] = useState<AuthState>("loading");
  const [error, setError] = useState<string | null>(null);
  const [username, setUsername] = useState<string | null>(null);
  const [avatar, setAvatar] = useState<string | null>(null);
  const [recentSaves, setRecentSaves] = useState<RecentSave[]>([]);
  const [isDark, setIsDark] = useState(false);

  useEffect(() => {
    Promise.all([
      getAuth(),
      getUsername(),
      getAvatar(),
      getRecentSaves(),
      getDarkMode(),
    ]).then(([auth, name, pic, saves, darkMode]) => {
      setAuthState(auth && Date.now() < auth.expiresAt ? "authenticated" : "unauthenticated");
      setUsername(name);
      setAvatar(pic);
      setRecentSaves(saves);
      setIsDark(darkMode);
    });
  }, []);

  async function handleLogin() {
    setError(null);
    try {
      await login();
      const [auth, name, pic] = await Promise.all([getAuth(), getUsername(), getAvatar()]);
      setAuthState(auth ? "authenticated" : "unauthenticated");
      setUsername(name);
      setAvatar(pic);
    } catch {
      setError("Login failed. Please try again.");
    }
  }

  async function handleLogout() {
    await clearAuth();
    setAuthState("unauthenticated");
  }

  async function handleToggleDark() {
    const next = !isDark;
    setIsDark(next);
    await setDarkMode(next);
  }

  function handleOpen() {
    const appUrl = import.meta.env.VITE_APP_URL as string;
    browser.tabs.create({ url: appUrl });
  }

  if (authState === "loading") {
    return <div style={{ width: 320, background: "#fff" }} />;
  }

  if (authState === "unauthenticated") {
    return <LoggedOut onLogin={handleLogin} error={error} />;
  }

  const c = isDark ? dark : light;
  return (
    <LoggedIn
      c={c}
      isDark={isDark}
      username={username}
      avatar={avatar}
      recentSaves={recentSaves}
      onToggleDark={handleToggleDark}
      onOpen={handleOpen}
      onLogout={handleLogout}
    />
  );
}

function LoggedOut({
  onLogin,
  error,
}: {
  onLogin: () => void;
  error: string | null;
}) {
  return (
    <div style={s.container}>
      {/* Header */}
      <div style={s.header}>
        <img src="../icons/icon48.png" width={16} height={16} alt="" />
        <span style={s.wordmark}>Bookleaf</span>
      </div>

      {/* Body */}
      <div style={s.loggedOutBody}>
        <div style={{ marginBottom: 16, opacity: 0.28 }}>
          <img src="../icons/icon48.png" width={38} height={38} alt="" />
        </div>
        <p style={s.tagline}>
          Save images to your Bookleaf collection as you browse the web.
        </p>
        {error && <p style={s.errorText}>{error}</p>}
        <button style={s.ctaButton} onClick={onLogin}>
          Log in to Bookleaf
        </button>
        <div style={s.signupLine}>
          New here?{" "}
          <span style={s.signupLink}>Sign up free</span>
        </div>
      </div>
    </div>
  );
}

function LoggedIn({
  c,
  isDark,
  username,
  avatar,
  recentSaves,
  onToggleDark,
  onOpen,
  onLogout,
}: {
  c: Colors;
  isDark: boolean;
  username: string | null;
  avatar: string | null;
  recentSaves: RecentSave[];
  onToggleDark: () => void;
  onOpen: () => void;
  onLogout: () => void;
}) {
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
          gap: 7,
        }}
      >
        <img src="../icons/icon48.png" width={16} height={16} alt="" />
        <span
          style={{
            fontSize: 14,
            fontWeight: 600,
            color: c.text,
            flex: 1,
            letterSpacing: "-0.02em",
          }}
        >
          Bookleaf
        </span>
        <button
          onClick={onOpen}
          style={{
            display: "flex",
            alignItems: "center",
            gap: 4,
            fontSize: 12,
            color: c.textSec,
            background: "transparent",
            border: `1px solid ${c.border}`,
            borderRadius: 6,
            padding: "4px 9px",
            cursor: "pointer",
            letterSpacing: "-0.01em",
          }}
        >
          Open{" "}
          <span style={{ fontSize: 9, display: "inline-block" }}>↗</span>
        </button>
      </div>

      {/* User row */}
      <div
        style={{
          height: 52,
          display: "flex",
          alignItems: "center",
          padding: "0 14px",
          borderBottom: `1px solid ${c.divider}`,
          gap: 10,
        }}
      >
        {avatar ? (
          <img
            src={avatar}
            width={28}
            height={28}
            alt=""
            style={{ borderRadius: "50%", flexShrink: 0, objectFit: "cover" }}
          />
        ) : (
          <div
            style={{
              width: 28,
              height: 28,
              borderRadius: "50%",
              flexShrink: 0,
              background:
                "linear-gradient(135deg,#f093fb 0%,#f5576c 50%,#4facfe 100%)",
            }}
          />
        )}
        <span
          style={{ fontSize: 13, fontWeight: 500, color: c.text, flex: 1 }}
        >
          {username ?? ""}
        </span>
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
          {isDark ? <SunIcon /> : <MoonIcon />}
        </button>
      </div>

      {/* Recently saved */}
      <div style={{ padding: "12px 14px 14px" }}>
        <div
          style={{ display: "flex", alignItems: "center", marginBottom: 9 }}
        >
          <span
            style={{
              fontSize: 10,
              fontWeight: 600,
              letterSpacing: "0.07em",
              textTransform: "uppercase",
              color: c.textSec,
              flex: 1,
            }}
          >
            Recently saved
          </span>
          {recentSaves.length > 0 && (
            <span
              style={{
                fontSize: 11,
                color: c.textSec,
                cursor: "pointer",
                textDecoration: "underline",
                textDecorationColor: c.textTer,
              }}
            >
              View all
            </span>
          )}
        </div>

        {recentSaves.length === 0 ? (
          <div
            style={{
              display: "flex",
              flexDirection: "column",
              alignItems: "center",
              justifyContent: "center",
              gap: 7,
              padding: "18px 0 14px",
            }}
          >
            <div style={{ opacity: 0.18 }}>
              <img src="../icons/icon48.png" width={28} height={28} alt="" />
            </div>
            <p
              style={{
                fontSize: 12,
                color: c.textSec,
                textAlign: "center",
                lineHeight: 1.6,
                margin: 0,
              }}
            >
              Nothing saved yet.
              <br />
              <span style={{ opacity: 0.7 }}>
                Right-click any image to save it.
              </span>
            </p>
          </div>
        ) : (
          <div style={{ display: "flex", gap: 6 }}>
            {recentSaves.map((save) => (
              <img
                key={save.imageId}
                src={save.dataUrl}
                title={save.title}
                alt={save.title}
                style={{
                  flex: 1,
                  aspectRatio: "1",
                  borderRadius: 7,
                  objectFit: "cover",
                  border: `1px solid ${c.thumbBorder}`,
                  cursor: "pointer",
                  minWidth: 0,
                }}
              />
            ))}
          </div>
        )}
      </div>

      {/* Footer */}
      <div
        style={{
          padding: "8px 14px 12px",
          borderTop: `1px solid ${c.divider}`,
          display: "flex",
          justifyContent: "flex-end",
        }}
      >
        <button
          onClick={onLogout}
          style={{
            fontSize: 12,
            color: c.textSec,
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

const s: Record<string, React.CSSProperties> = {
  container: {
    width: 320,
    background: "#fff",
    border: "1px solid #e0e0e0",
    fontFamily:
      '-apple-system, BlinkMacSystemFont, "Segoe UI", "Helvetica Neue", sans-serif',
    overflow: "hidden",
  },
  header: {
    height: 44,
    display: "flex",
    alignItems: "center",
    padding: "0 14px",
    borderBottom: "1px solid #e8e8e8",
    gap: 7,
  },
  wordmark: {
    fontSize: 14,
    fontWeight: 600,
    color: "#1c1c1c",
    letterSpacing: "-0.02em",
  },
  loggedOutBody: {
    padding: "28px 18px 24px",
    display: "flex",
    flexDirection: "column",
    alignItems: "center",
  },
  tagline: {
    fontSize: 13,
    color: "#888",
    lineHeight: 1.65,
    textAlign: "center",
    marginBottom: 18,
    margin: "0 0 18px",
  },
  errorText: {
    margin: "0 0 10px",
    fontSize: 13,
    color: "#c0392b",
    textAlign: "center",
  },
  ctaButton: {
    width: "100%",
    padding: "10px 0",
    background: "#1c1c1c",
    color: "#fff",
    border: "none",
    borderRadius: 8,
    fontSize: 13,
    fontWeight: 500,
    cursor: "pointer",
    marginBottom: 11,
    letterSpacing: "-0.01em",
  },
  signupLine: {
    fontSize: 12,
    color: "#bbb",
  },
  signupLink: {
    color: "#1c1c1c",
    fontWeight: 500,
    cursor: "pointer",
  },
};

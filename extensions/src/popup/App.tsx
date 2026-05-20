import { useEffect, useState } from "react";
import { login } from "../lib/auth";
import { clearAuth, getAuth } from "../lib/storage";

type AuthState = "loading" | "unauthenticated" | "authenticated";

export default function App() {
  const [authState, setAuthState] = useState<AuthState>("loading");
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    getAuth().then((auth) => {
      setAuthState(auth ? "authenticated" : "unauthenticated");
    });
  }, []);

  async function handleLogin() {
    setError(null);
    try {
      await login();
      setAuthState("authenticated");
    } catch {
      setError("Login failed. Please try again.");
    }
  }

  async function handleLogout() {
    await clearAuth();
    setAuthState("unauthenticated");
  }

  if (authState === "loading") {
    return <div style={styles.container} />;
  }

  return (
    <div style={styles.container}>
      <h2 style={styles.title}>Bookleaf</h2>
      {authState === "authenticated" ? (
        <>
          <p style={styles.status}>Logged in</p>
          <button style={styles.button} onClick={handleLogout}>
            Logout
          </button>
        </>
      ) : (
        <>
          {error && <p style={styles.error}>{error}</p>}
          <button style={styles.button} onClick={handleLogin}>
            Login with Bookleaf
          </button>
        </>
      )}
    </div>
  );
}

const styles: Record<string, React.CSSProperties> = {
  container: {
    padding: "16px",
    minHeight: "80px",
    display: "flex",
    flexDirection: "column",
    gap: "8px",
  },
  title: {
    margin: 0,
    fontSize: "16px",
    fontWeight: 600,
  },
  status: {
    margin: 0,
    fontSize: "14px",
    color: "#555",
  },
  button: {
    padding: "8px 12px",
    fontSize: "14px",
    cursor: "pointer",
    borderRadius: "4px",
    border: "1px solid #ccc",
    background: "#fff",
  },
  error: {
    margin: 0,
    fontSize: "13px",
    color: "#c0392b",
  },
};

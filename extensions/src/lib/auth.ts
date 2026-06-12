import browser from "webextension-polyfill";
import { setAuth, setUsername, setAvatar, type BookleafAuth } from "./storage";

function generateCodeVerifier(): string {
  const array = new Uint8Array(32);
  crypto.getRandomValues(array);
  return btoa(String.fromCharCode(...array))
    .replace(/\+/g, "-")
    .replace(/\//g, "_")
    .replace(/=/g, "");
}

async function generateCodeChallenge(verifier: string): Promise<string> {
  const encoder = new TextEncoder();
  const data = encoder.encode(verifier);
  const digest = await crypto.subtle.digest("SHA-256", data);
  return btoa(String.fromCharCode(...new Uint8Array(digest)))
    .replace(/\+/g, "-")
    .replace(/\//g, "_")
    .replace(/=/g, "");
}

function getRedirectUri(): string {
  return browser.identity.getRedirectURL();
}

function generateState(): string {
  const array = new Uint8Array(16);
  crypto.getRandomValues(array);
  return btoa(String.fromCharCode(...array))
    .replace(/\+/g, "-")
    .replace(/\//g, "_")
    .replace(/=/g, "");
}

function buildAuthUrl(codeChallenge: string): string {
  const issuerUrl = import.meta.env.VITE_KINDE_ISSUER_URL;
  const clientId = import.meta.env.VITE_KINDE_CLIENT_ID;
  const redirectUri = getRedirectUri();

  const audience = import.meta.env.VITE_KINDE_AUDIENCE;

  const params = new URLSearchParams({
    response_type: "code",
    client_id: clientId,
    redirect_uri: redirectUri,
    scope: "openid profile email",
    code_challenge: codeChallenge,
    code_challenge_method: "S256",
    state: generateState(),
    ...(audience && { audience }),
  });

  return `${issuerUrl}/oauth2/auth?${params.toString()}`;
}

export function decodeJwtPayload(token: string): Record<string, unknown> {
  const parts = token.split(".");
  if (parts.length < 2) return {};
  const payload = parts[1].replace(/-/g, "+").replace(/_/g, "/");
  try {
    return JSON.parse(atob(payload)) as Record<string, unknown>;
  } catch {
    return {};
  }
}

async function exchangeCodeForTokens(
  code: string,
  codeVerifier: string,
): Promise<{ auth: BookleafAuth; idToken: string | null }> {
  const issuerUrl = import.meta.env.VITE_KINDE_ISSUER_URL;
  const clientId = import.meta.env.VITE_KINDE_CLIENT_ID;
  const redirectUri = getRedirectUri();

  const response = await fetch(`${issuerUrl}/oauth2/token`, {
    method: "POST",
    headers: { "Content-Type": "application/x-www-form-urlencoded" },
    body: new URLSearchParams({
      grant_type: "authorization_code",
      code,
      code_verifier: codeVerifier,
      client_id: clientId,
      redirect_uri: redirectUri,
    }),
  });

  if (!response.ok) {
    throw new Error(`Token exchange failed: ${response.status}`);
  }

  const json = (await response.json()) as {
    access_token: string;
    id_token?: string;
    refresh_token?: string;
    expires_in: number;
  };

  const auth: BookleafAuth = {
    accessToken: json.access_token,
    ...(json.refresh_token && { refreshToken: json.refresh_token }),
    expiresAt: Date.now() + json.expires_in * 1000,
  };

  return { auth, idToken: json.id_token ?? null };
}

export async function login(): Promise<void> {
  const codeVerifier = generateCodeVerifier();
  const codeChallenge = await generateCodeChallenge(codeVerifier);
  const authUrl = buildAuthUrl(codeChallenge);

  const redirectUrl = await browser.identity.launchWebAuthFlow({
    url: authUrl,
    interactive: true,
  });

  if (!redirectUrl) return;

  const url = new URL(redirectUrl);
  const code = url.searchParams.get("code");
  if (!code) return;

  const { auth, idToken } = await exchangeCodeForTokens(code, codeVerifier);
  await setAuth(auth);

  if (idToken) {
    const payload = decodeJwtPayload(idToken);
    const username =
      (payload.given_name as string) ||
      (payload.name as string) ||
      (payload.email as string) ||
      null;
    if (username) await setUsername(username);
    const picture = payload.picture as string | undefined;
    if (picture) await setAvatar(picture);
  }
}

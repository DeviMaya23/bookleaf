import { beforeEach, describe, expect, it, vi } from "vitest";
import browser from "webextension-polyfill";
import { setAuth, setAvatar, setUsername } from "./storage";

vi.mock("webextension-polyfill", async () => {
  const { createBrowserMock } = await import("../test/browserMock");
  return { default: createBrowserMock() };
});
vi.mock("./storage", () => ({
  setAuth: vi.fn(),
  setUsername: vi.fn(),
  setAvatar: vi.fn(),
}));

import { buildAuthUrl, decodeJwtPayload, exchangeCodeForTokens, login } from "./auth";

const browserMock = browser as unknown as {
  identity: { launchWebAuthFlow: ReturnType<typeof vi.fn> };
};

function base64UrlEncode(json: unknown): string {
  return btoa(JSON.stringify(json)).replace(/\+/g, "-").replace(/\//g, "_").replace(/=/g, "");
}

describe("decodeJwtPayload", () => {
  it("decodes a valid token payload", () => {
    const payload = { email: "user@example.com" };
    const token = `header.${base64UrlEncode(payload)}.signature`;
    expect(decodeJwtPayload(token)).toEqual(payload);
  });

  it("returns an empty object for a malformed token", () => {
    expect(decodeJwtPayload("not-a-jwt")).toEqual({});
  });

  it("returns an empty object when the payload is not valid JSON", () => {
    const token = `header.${btoa("not json")}.signature`;
    expect(decodeJwtPayload(token)).toEqual({});
  });
});

describe("buildAuthUrl", () => {
  beforeEach(() => {
    vi.stubEnv("VITE_KINDE_ISSUER_URL", "https://issuer.test");
    vi.stubEnv("VITE_KINDE_CLIENT_ID", "client-123");
  });

  it("includes the audience param when configured", () => {
    vi.stubEnv("VITE_KINDE_AUDIENCE", "https://api.test");
    const url = new URL(buildAuthUrl("challenge"));
    expect(url.searchParams.get("audience")).toBe("https://api.test");
  });

  it("omits the audience param when not configured", () => {
    vi.stubEnv("VITE_KINDE_AUDIENCE", "");
    const url = new URL(buildAuthUrl("challenge"));
    expect(url.searchParams.has("audience")).toBe(false);
  });
});

describe("exchangeCodeForTokens", () => {
  beforeEach(() => {
    vi.stubEnv("VITE_KINDE_ISSUER_URL", "https://issuer.test");
    vi.stubEnv("VITE_KINDE_CLIENT_ID", "client-123");
  });

  it("throws when the token response is not OK", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => new Response(null, { status: 400 })),
    );

    await expect(exchangeCodeForTokens("code", "verifier")).rejects.toThrow(
      "Token exchange failed: 400",
    );
  });

  it("returns the auth and id token on success", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(
        async () =>
          new Response(
            JSON.stringify({
              access_token: "access-123",
              id_token: "id-123",
              expires_in: 3600,
            }),
            { status: 200 },
          ),
      ),
    );

    const result = await exchangeCodeForTokens("code", "verifier");

    expect(result.auth.accessToken).toBe("access-123");
    expect(result.idToken).toBe("id-123");
  });
});

describe("login", () => {
  beforeEach(() => {
    vi.stubEnv("VITE_KINDE_ISSUER_URL", "https://issuer.test");
    vi.stubEnv("VITE_KINDE_CLIENT_ID", "client-123");
    browserMock.identity.launchWebAuthFlow.mockResolvedValue(
      "https://redirect.test?code=auth-code",
    );
    vi.stubGlobal(
      "fetch",
      vi.fn(
        async () =>
          new Response(
            JSON.stringify({ access_token: "access-123", expires_in: 3600 }),
            { status: 200 },
          ),
      ),
    );
  });

  it("falls back through given_name, name, email for the username", async () => {
    const idToken = `header.${base64UrlEncode({ name: "Jane" })}.sig`;
    vi.mocked(fetch).mockResolvedValue(
      new Response(
        JSON.stringify({ access_token: "access-123", id_token: idToken, expires_in: 3600 }),
        { status: 200 },
      ),
    );

    await login();

    expect(vi.mocked(setUsername)).toHaveBeenCalledWith("Jane");
  });

  it("calls setAvatar only when the id token carries a picture claim", async () => {
    const idToken = `header.${base64UrlEncode({ email: "user@example.com", picture: "https://avatar.test/p.png" })}.sig`;
    vi.mocked(fetch).mockResolvedValue(
      new Response(
        JSON.stringify({ access_token: "access-123", id_token: idToken, expires_in: 3600 }),
        { status: 200 },
      ),
    );

    await login();

    expect(vi.mocked(setAvatar)).toHaveBeenCalledWith("https://avatar.test/p.png");
    expect(vi.mocked(setAuth)).toHaveBeenCalled();
  });
});

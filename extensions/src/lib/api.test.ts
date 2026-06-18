import { beforeEach, describe, expect, it, vi } from "vitest";
import { apiFetch } from "./api";
import { getAuth } from "./storage";

vi.mock("./storage", () => ({
  getAuth: vi.fn(),
}));

describe("apiFetch", () => {
  beforeEach(() => {
    vi.stubGlobal("fetch", vi.fn(async () => new Response()));
    vi.stubEnv("VITE_API_BASE_URL", "https://api.bookleaf.test");
  });

  it("sets the Authorization header when auth is present", async () => {
    vi.mocked(getAuth).mockResolvedValue({
      accessToken: "token-123",
      expiresAt: Date.now() + 1000,
    });

    await apiFetch("/images");

    const [, options] = vi.mocked(fetch).mock.calls[0];
    const headers = new Headers(options?.headers);
    expect(headers.get("Authorization")).toBe("Bearer token-123");
  });

  it("omits the Authorization header when auth is absent", async () => {
    vi.mocked(getAuth).mockResolvedValue(null);

    await apiFetch("/images");

    const [, options] = vi.mocked(fetch).mock.calls[0];
    const headers = new Headers(options?.headers);
    expect(headers.has("Authorization")).toBe(false);
  });
});

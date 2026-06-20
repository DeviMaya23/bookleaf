import { beforeEach, describe, expect, it, vi } from "vitest";
import browser from "webextension-polyfill";
import { apiFetch } from "../lib/api";
import { resolveHighResReferrer, resolveHighResUrl, validateCandidate } from "../lib/highResFetch";
import { addRecentSave, getAuth } from "../lib/storage";

vi.mock("webextension-polyfill", async () => {
  const { createBrowserMock } = await import("../test/browserMock");
  return { default: createBrowserMock() };
});
vi.mock("../lib/api", () => ({ apiFetch: vi.fn() }));
vi.mock("../lib/highResFetch", () => ({
  resolveHighResUrl: vi.fn(),
  resolveHighResReferrer: vi.fn(),
  validateCandidate: vi.fn(),
}));
vi.mock("../lib/storage", () => ({ getAuth: vi.fn(), addRecentSave: vi.fn() }));

import {
  blobToDataUrl,
  handleContextMenuClick,
  handleSave,
  isTokenValid,
  resolveImageBlob,
  resolvedContextByTab,
  saveImage,
} from "./index";

function jsonResponse(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), { status });
}

describe("isTokenValid", () => {
  it("is valid when the token is not expired", () => {
    expect(isTokenValid({ accessToken: "t", expiresAt: Date.now() + 10_000 })).toBe(true);
  });

  it("is invalid when the token is expired", () => {
    expect(isTokenValid({ accessToken: "t", expiresAt: Date.now() - 10_000 })).toBe(false);
  });

  it("is invalid when auth is null", () => {
    expect(isTokenValid(null)).toBe(false);
  });
});

describe("blobToDataUrl", () => {
  it("produces the expected data URL for a known blob", async () => {
    const blob = new Blob([new Uint8Array([72, 105])]);
    expect(await blobToDataUrl(blob)).toBe("data:image/jpeg;base64,SGk=");
  });
});

describe("resolveImageBlob", () => {
  beforeEach(() => {
    vi.stubGlobal("fetch", vi.fn());
  });

  it("returns the high-res candidate when valid", async () => {
    vi.mocked(resolveHighResUrl).mockReturnValue("https://hi-res.test/img.jpg");
    const candidateResponse = new Response(new Blob([new Uint8Array([1])]));
    vi.mocked(fetch).mockResolvedValue(candidateResponse);
    vi.mocked(validateCandidate).mockResolvedValue({ valid: true, bitmap: null });

    const result = await resolveImageBlob("https://example.com/img.jpg");

    expect(fetch).toHaveBeenCalledWith("https://hi-res.test/img.jpg");
    expect(result.bitmap).toBeNull();
  });

  it("falls back to the original srcUrl when candidate validation fails", async () => {
    vi.mocked(resolveHighResUrl).mockReturnValue("https://hi-res.test/img.jpg");
    const candidateResponse = new Response(new Blob([new Uint8Array([1])]));
    const fallbackResponse = new Response(new Blob([new Uint8Array([2])]), {
      status: 200,
      headers: { "Content-Type": "image/png" },
    });
    vi.mocked(fetch)
      .mockResolvedValueOnce(candidateResponse)
      .mockResolvedValueOnce(fallbackResponse);
    vi.mocked(validateCandidate).mockResolvedValue({ valid: false, bitmap: null });

    const result = await resolveImageBlob("https://example.com/img.jpg");

    expect(fetch).toHaveBeenCalledWith("https://example.com/img.jpg");
    expect(result.mimeType).toBe("image/png");
  });
});

describe("saveImage", () => {
  const baseArgs = {
    blob: new Blob([new Uint8Array([1])]),
    mimeType: "image/jpeg",
    title: "Title",
    pageUrl: "https://example.com",
    accessToken: "token",
  };

  beforeEach(() => {
    vi.stubGlobal("fetch", vi.fn());
  });

  it("assembles image_id from init, parallel PUTs, and complete", async () => {
    vi.mocked(apiFetch)
      .mockResolvedValueOnce(
        jsonResponse({
          upload_url: "https://r2.test/upload",
          thumbnail_upload_url: "https://r2.test/thumb",
          id: "img-1",
        }),
      )
      .mockResolvedValueOnce(jsonResponse({}));
    vi.mocked(fetch).mockResolvedValue(new Response(null, { status: 200 }));

    const result = await saveImage(baseArgs);

    expect(result).toBe("img-1");
  });

  it("throws when the init request fails", async () => {
    vi.mocked(apiFetch).mockResolvedValueOnce(new Response(null, { status: 500 }));

    await expect(saveImage(baseArgs)).rejects.toThrow("POST /images failed: 500");
  });

  it("throws when a PUT upload fails", async () => {
    vi.mocked(apiFetch).mockResolvedValueOnce(
      jsonResponse({
        upload_url: "https://r2.test/upload",
        thumbnail_upload_url: "https://r2.test/thumb",
        id: "img-1",
      }),
    );
    vi.mocked(fetch).mockResolvedValue(new Response(null, { status: 500 }));

    await expect(saveImage(baseArgs)).rejects.toThrow("PUT to R2 failed: 500");
  });

  it("throws when the complete request fails", async () => {
    vi.mocked(apiFetch)
      .mockResolvedValueOnce(
        jsonResponse({
          upload_url: "https://r2.test/upload",
          thumbnail_upload_url: "https://r2.test/thumb",
          id: "img-1",
        }),
      )
      .mockResolvedValueOnce(new Response(null, { status: 500 }));
    vi.mocked(fetch).mockResolvedValue(new Response(null, { status: 200 }));

    await expect(saveImage(baseArgs)).rejects.toThrow("POST /complete failed: 500");
  });
});

describe("handleSave", () => {
  beforeEach(() => {
    vi.stubGlobal("fetch", vi.fn());
    vi.mocked(resolveHighResUrl).mockReturnValue(null);
  });

  it("sends an error toast and returns early when auth is invalid", async () => {
    vi.mocked(getAuth).mockResolvedValue(null);

    await handleSave({ srcUrl: "https://example.com/img.jpg", pageUrl: "", title: "t", tabId: 1 });

    expect(browser.tabs.sendMessage).toHaveBeenCalledWith(
      1,
      expect.objectContaining({ variant: "error", title: "Bookleaf" }),
    );
    expect(apiFetch).not.toHaveBeenCalled();
  });

  it("sends a success toast and records the recent save on success", async () => {
    vi.mocked(getAuth).mockResolvedValue({ accessToken: "token", expiresAt: Date.now() + 10_000 });
    vi.mocked(fetch).mockResolvedValue(
      new Response(new Blob([new Uint8Array([1])]), {
        status: 200,
        headers: { "Content-Type": "image/jpeg" },
      }),
    );
    vi.mocked(apiFetch)
      .mockResolvedValueOnce(
        jsonResponse({
          upload_url: "https://r2.test/upload",
          thumbnail_upload_url: "https://r2.test/thumb",
          id: "img-1",
        }),
      )
      .mockResolvedValueOnce(jsonResponse({}));

    await handleSave({ srcUrl: "https://example.com/img.jpg", pageUrl: "", title: "t", tabId: 1 });

    expect(browser.tabs.sendMessage).toHaveBeenCalledWith(
      1,
      expect.objectContaining({ variant: "success" }),
    );
    expect(addRecentSave).toHaveBeenCalledWith(
      expect.objectContaining({ imageId: "img-1", title: "t" }),
    );
  });

  it("sends an error toast and skips the recent save when saving throws", async () => {
    vi.mocked(getAuth).mockResolvedValue({ accessToken: "token", expiresAt: Date.now() + 10_000 });
    vi.mocked(fetch).mockResolvedValue(
      new Response(new Blob([new Uint8Array([1])]), {
        status: 200,
        headers: { "Content-Type": "image/jpeg" },
      }),
    );
    vi.mocked(apiFetch).mockResolvedValueOnce(new Response(null, { status: 500 }));

    await handleSave({ srcUrl: "https://example.com/img.jpg", pageUrl: "", title: "t", tabId: 1 });

    expect(browser.tabs.sendMessage).toHaveBeenCalledWith(
      1,
      expect.objectContaining({ variant: "error", title: "Couldn't save image." }),
    );
    expect(addRecentSave).not.toHaveBeenCalled();
  });
});

describe("handleContextMenuClick", () => {
  const tab = { id: 1, title: "t" } as browser.Tabs.Tab;

  beforeEach(() => {
    resolvedContextByTab.clear();
    vi.stubGlobal("fetch", vi.fn());
    vi.mocked(resolveHighResUrl).mockReturnValue(null);
    vi.mocked(getAuth).mockResolvedValue({ accessToken: "token", expiresAt: Date.now() + 10_000 });
    vi.mocked(fetch).mockResolvedValue(
      new Response(new Blob([new Uint8Array([1])]), {
        status: 200,
        headers: { "Content-Type": "image/jpeg" },
      }),
    );
    vi.mocked(apiFetch)
      .mockResolvedValueOnce(
        jsonResponse({
          upload_url: "https://r2.test/upload",
          thumbnail_upload_url: "https://r2.test/thumb",
          id: "img-1",
        }),
      )
      .mockResolvedValueOnce(jsonResponse({}));
  });

  it("overrides source_url with linkUrl for image-context when a permalink rule matches", async () => {
    handleContextMenuClick(
      {
        menuItemId: "save-to-bookleaf",
        srcUrl: "https://pbs.twimg.com/media/img.jpg",
        pageUrl: "https://x.com/username",
        linkUrl: "https://x.com/username/status/123456789",
      } as browser.Menus.OnClickData,
      tab,
    );
    await vi.waitFor(() => expect(apiFetch).toHaveBeenCalled());

    expect(apiFetch).toHaveBeenCalledWith(
      "/images",
      expect.objectContaining({
        body: expect.stringContaining('"source_url":"https://x.com/username/status/123456789"'),
      }),
    );
  });

  it("keeps pageUrl as source_url for image-context when linkUrl matches no rule", async () => {
    handleContextMenuClick(
      {
        menuItemId: "save-to-bookleaf",
        srcUrl: "https://example.com/img.jpg",
        pageUrl: "https://example.com/page",
        linkUrl: "https://example.com/unrelated",
      } as browser.Menus.OnClickData,
      tab,
    );
    await vi.waitFor(() => expect(apiFetch).toHaveBeenCalled());

    expect(apiFetch).toHaveBeenCalledWith(
      "/images",
      expect.objectContaining({
        body: expect.stringContaining('"source_url":"https://example.com/page"'),
      }),
    );
  });

  it("uses the resolved srcUrl and linkUrl as source_url for link-context", async () => {
    resolvedContextByTab.set(1, { srcUrl: "https://i.pinimg.com/originals/img.jpg" });

    handleContextMenuClick(
      {
        menuItemId: "save-to-bookleaf-link",
        pageUrl: "https://www.pinterest.com/feed",
        linkUrl: "https://www.pinterest.com/pin/123",
      } as browser.Menus.OnClickData,
      tab,
    );
    await vi.waitFor(() => expect(apiFetch).toHaveBeenCalled());

    expect(fetch).toHaveBeenCalledWith("https://i.pinimg.com/originals/img.jpg");
    expect(apiFetch).toHaveBeenCalledWith(
      "/images",
      expect.objectContaining({
        body: expect.stringContaining('"source_url":"https://www.pinterest.com/pin/123"'),
      }),
    );
  });

  it("fails gracefully with no upload when no resolved srcUrl exists for link-context", async () => {
    handleContextMenuClick(
      {
        menuItemId: "save-to-bookleaf-link",
        pageUrl: "https://www.pinterest.com/feed",
        linkUrl: "https://www.pinterest.com/pin/123",
      } as browser.Menus.OnClickData,
      tab,
    );
    await vi.waitFor(() =>
      expect(browser.tabs.sendMessage).toHaveBeenCalledWith(
        1,
        expect.objectContaining({ variant: "error", title: "Couldn't save image." }),
      ),
    );

    expect(apiFetch).not.toHaveBeenCalled();
  });

  it("uses '@handle: tweet text...' as title when tweet text was resolved", async () => {
    resolvedContextByTab.set(1, { title: "a".repeat(150) });

    handleContextMenuClick(
      {
        menuItemId: "save-to-bookleaf",
        srcUrl: "https://pbs.twimg.com/media/img.jpg",
        pageUrl: "https://x.com/home",
        linkUrl: "https://x.com/username/status/123456789",
      } as browser.Menus.OnClickData,
      tab,
    );
    await vi.waitFor(() => expect(apiFetch).toHaveBeenCalled());

    expect(apiFetch).toHaveBeenCalledWith(
      "/images",
      expect.objectContaining({
        body: expect.stringContaining(`"title":"@username: ${"a".repeat(100)}..."`),
      }),
    );
  });

  it("appends '...' even when the resolved tweet text is under 100 characters", async () => {
    resolvedContextByTab.set(1, { title: "short tweet" });

    handleContextMenuClick(
      {
        menuItemId: "save-to-bookleaf",
        srcUrl: "https://pbs.twimg.com/media/img.jpg",
        pageUrl: "https://x.com/home",
        linkUrl: "https://x.com/username/status/123456789",
      } as browser.Menus.OnClickData,
      tab,
    );
    await vi.waitFor(() => expect(apiFetch).toHaveBeenCalled());

    expect(apiFetch).toHaveBeenCalledWith(
      "/images",
      expect.objectContaining({
        body: expect.stringContaining('"title":"@username: short tweet..."'),
      }),
    );
  });

  it("uses '@handle' alone as title when no tweet text was resolved", async () => {
    handleContextMenuClick(
      {
        menuItemId: "save-to-bookleaf",
        srcUrl: "https://pbs.twimg.com/media/img.jpg",
        pageUrl: "https://x.com/home",
        linkUrl: "https://x.com/username/status/123456789",
      } as browser.Menus.OnClickData,
      tab,
    );
    await vi.waitFor(() => expect(apiFetch).toHaveBeenCalled());

    expect(apiFetch).toHaveBeenCalledWith(
      "/images",
      expect.objectContaining({
        body: expect.stringContaining('"title":"@username"'),
      }),
    );
  });

  it("uses tab.title for non-Twitter saves, unaffected by this change", async () => {
    handleContextMenuClick(
      {
        menuItemId: "save-to-bookleaf",
        srcUrl: "https://example.com/img.jpg",
        pageUrl: "https://example.com/page",
        linkUrl: "https://example.com/unrelated",
      } as browser.Menus.OnClickData,
      tab,
    );
    await vi.waitFor(() => expect(apiFetch).toHaveBeenCalled());

    expect(apiFetch).toHaveBeenCalledWith(
      "/images",
      expect.objectContaining({
        body: expect.stringContaining('"title":"t"'),
      }),
    );
  });
});

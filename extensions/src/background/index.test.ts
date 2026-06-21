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
  handleCapture,
  handleContextMenuClick,
  handleDragSaveMessage,
  handleSave,
  handleSnipCapturedMessage,
  handleSnipCommand,
  isTokenValid,
  persistImage,
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

describe("persistImage", () => {
  const baseArgs = {
    blob: new Blob([new Uint8Array([1])]),
    mimeType: "image/png",
    title: "Snip",
    pageUrl: "https://example.com/page",
    tabId: 1,
  };

  beforeEach(() => {
    vi.stubGlobal("fetch", vi.fn());
  });

  it("sends a success toast and records the recent save when auth is valid and upload succeeds", async () => {
    vi.mocked(getAuth).mockResolvedValue({ accessToken: "token", expiresAt: Date.now() + 10_000 });
    vi.mocked(fetch).mockResolvedValue(new Response(null, { status: 200 }));
    vi.mocked(apiFetch)
      .mockResolvedValueOnce(
        jsonResponse({
          upload_url: "https://r2.test/upload",
          thumbnail_upload_url: "https://r2.test/thumb",
          id: "img-1",
        }),
      )
      .mockResolvedValueOnce(jsonResponse({}));

    await persistImage(baseArgs);

    expect(browser.tabs.sendMessage).toHaveBeenCalledWith(
      1,
      expect.objectContaining({ variant: "success" }),
    );
    expect(addRecentSave).toHaveBeenCalledWith(
      expect.objectContaining({ imageId: "img-1", title: "Snip" }),
    );
  });

  it("sends a login-required toast and attempts no upload when there is no valid token", async () => {
    vi.mocked(getAuth).mockResolvedValue(null);

    await persistImage(baseArgs);

    expect(browser.tabs.sendMessage).toHaveBeenCalledWith(
      1,
      expect.objectContaining({ variant: "error", title: "Bookleaf" }),
    );
    expect(apiFetch).not.toHaveBeenCalled();
  });

  it("sends an error toast and skips the recent save when an upload step fails", async () => {
    vi.mocked(getAuth).mockResolvedValue({ accessToken: "token", expiresAt: Date.now() + 10_000 });
    vi.mocked(apiFetch).mockResolvedValueOnce(new Response(null, { status: 500 }));

    await persistImage(baseArgs);

    expect(browser.tabs.sendMessage).toHaveBeenCalledWith(
      1,
      expect.objectContaining({ variant: "error", title: "Couldn't save image." }),
    );
    expect(addRecentSave).not.toHaveBeenCalled();
  });
});

describe("handleCapture", () => {
  beforeEach(() => {
    vi.stubGlobal("fetch", vi.fn());
    vi.mocked(getAuth).mockResolvedValue({ accessToken: "token", expiresAt: Date.now() + 10_000 });
    vi.mocked(fetch).mockResolvedValue(new Response(null, { status: 200 }));
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

  it("delegates to persistImage with the given blob/mimeType/title/pageUrl, with no srcUrl fetch attempted", async () => {
    const blob = new Blob([new Uint8Array([1])]);

    await handleCapture({
      blob,
      mimeType: "image/png",
      pageUrl: "https://example.com/page",
      title: "Tab Title",
      tabId: 1,
    });

    expect(fetch).not.toHaveBeenCalledWith("https://example.com/page");
    expect(apiFetch).toHaveBeenCalledWith(
      "/images",
      expect.objectContaining({
        body: expect.stringContaining('"source_url":"https://example.com/page"'),
      }),
    );
    expect(addRecentSave).toHaveBeenCalledWith(expect.objectContaining({ title: "Tab Title" }));
  });
});

describe("handleSnipCommand", () => {
  it("captures the active tab and sends the frame to its content script", async () => {
    vi.mocked(browser.tabs.query).mockResolvedValue([{ id: 1 } as browser.Tabs.Tab]);
    vi.mocked(browser.tabs.captureVisibleTab).mockResolvedValue("data:image/png;base64,abc");

    await handleSnipCommand("snip-capture");

    expect(browser.tabs.sendMessage).toHaveBeenCalledWith(1, {
      type: "snip-frame",
      dataUrl: "data:image/png;base64,abc",
    });
  });

  it("ignores commands other than snip-capture", async () => {
    await handleSnipCommand("some-other-command");

    expect(browser.tabs.captureVisibleTab).not.toHaveBeenCalled();
  });

  it("does nothing when there is no active tab", async () => {
    vi.mocked(browser.tabs.query).mockResolvedValue([]);

    await handleSnipCommand("snip-capture");

    expect(browser.tabs.captureVisibleTab).not.toHaveBeenCalled();
  });

  it("resolves cleanly when the content script can't be reached (e.g. a restricted page)", async () => {
    vi.mocked(browser.tabs.query).mockResolvedValue([{ id: 1 } as browser.Tabs.Tab]);
    vi.mocked(browser.tabs.captureVisibleTab).mockResolvedValue("data:image/png;base64,abc");
    vi.mocked(browser.tabs.sendMessage).mockRejectedValueOnce(
      new Error("Receiving end does not exist."),
    );

    await expect(handleSnipCommand("snip-capture")).resolves.toBeUndefined();
  });
});

describe("handleSnipCapturedMessage", () => {
  beforeEach(() => {
    vi.stubGlobal("fetch", vi.fn());
    vi.mocked(getAuth).mockResolvedValue({ accessToken: "token", expiresAt: Date.now() + 10_000 });
    vi.mocked(fetch).mockResolvedValue(new Response(null, { status: 200 }));
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

  it("uses tab.url and tab.title unmodified, with no per-site resolution", async () => {
    const tab = {
      id: 1,
      url: "https://x.com/username",
      title: "A tweet, by someone",
    } as browser.Tabs.Tab;

    handleSnipCapturedMessage(
      { type: "snip-captured", blob: new Blob([new Uint8Array([1])]), mimeType: "image/png" },
      tab,
    );
    await vi.waitFor(() => expect(apiFetch).toHaveBeenCalled());

    expect(apiFetch).toHaveBeenCalledWith(
      "/images",
      expect.objectContaining({
        body: expect.stringContaining('"title":"A tweet, by someone"'),
      }),
    );
    expect(apiFetch).toHaveBeenCalledWith(
      "/images",
      expect.objectContaining({
        body: expect.stringContaining('"source_url":"https://x.com/username"'),
      }),
    );
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

  it("uses the resolved alt text verbatim for an Imgur save", async () => {
    resolvedContextByTab.set(1, { title: "a cute cat" });

    handleContextMenuClick(
      {
        menuItemId: "save-to-bookleaf",
        srcUrl: "https://i.imgur.com/abc123.jpg",
        pageUrl: "https://imgur.com/gallery/abc123",
      } as browser.Menus.OnClickData,
      tab,
    );
    await vi.waitFor(() => expect(apiFetch).toHaveBeenCalled());

    expect(apiFetch).toHaveBeenCalledWith(
      "/images",
      expect.objectContaining({
        body: expect.stringContaining('"title":"a cute cat"'),
      }),
    );
  });

  it("falls back to tab.title when no resolved title exists for an Imgur save", async () => {
    handleContextMenuClick(
      {
        menuItemId: "save-to-bookleaf",
        srcUrl: "https://i.imgur.com/abc123.jpg",
        pageUrl: "https://imgur.com/gallery/abc123",
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

  it("uses the resolved alt text verbatim for an Instagram save with full caption text", async () => {
    resolvedContextByTab.set(1, { title: "A beautiful sunset over the mountains" });

    handleContextMenuClick(
      {
        menuItemId: "save-to-bookleaf",
        srcUrl: "https://instagram.com/p/abc123/media.jpg",
        pageUrl: "https://www.instagram.com/p/abc123",
      } as browser.Menus.OnClickData,
      tab,
    );
    await vi.waitFor(() => expect(apiFetch).toHaveBeenCalled());

    expect(apiFetch).toHaveBeenCalledWith(
      "/images",
      expect.objectContaining({
        body: expect.stringContaining('"title":"A beautiful sunset over the mountains"'),
      }),
    );
  });

  it("uses the resolved alt text verbatim for an Instagram save with the generated 'Photo by' form", async () => {
    resolvedContextByTab.set(1, { title: "Photo by username on January 1, 2024." });

    handleContextMenuClick(
      {
        menuItemId: "save-to-bookleaf",
        srcUrl: "https://instagram.com/p/abc123/media.jpg",
        pageUrl: "https://www.instagram.com/p/abc123",
      } as browser.Menus.OnClickData,
      tab,
    );
    await vi.waitFor(() => expect(apiFetch).toHaveBeenCalled());

    expect(apiFetch).toHaveBeenCalledWith(
      "/images",
      expect.objectContaining({
        body: expect.stringContaining('"title":"Photo by username on January 1, 2024."'),
      }),
    );
  });

  it("falls back to tab.title when no resolved title exists for an Instagram save", async () => {
    handleContextMenuClick(
      {
        menuItemId: "save-to-bookleaf",
        srcUrl: "https://instagram.com/p/abc123/media.jpg",
        pageUrl: "https://www.instagram.com/p/abc123",
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

  it("uses the resolved alt/aria-label text verbatim for a Facebook save", async () => {
    resolvedContextByTab.set(1, { title: "May be an image of gelato and text" });

    handleContextMenuClick(
      {
        menuItemId: "save-to-bookleaf",
        srcUrl: "https://scontent.facebook.com/abc123.jpg",
        pageUrl: "https://www.facebook.com/username/posts/123",
      } as browser.Menus.OnClickData,
      tab,
    );
    await vi.waitFor(() => expect(apiFetch).toHaveBeenCalled());

    expect(apiFetch).toHaveBeenCalledWith(
      "/images",
      expect.objectContaining({
        body: expect.stringContaining('"title":"May be an image of gelato and text"'),
      }),
    );
  });

  it("falls back to tab.title when no resolved title exists for a Facebook save", async () => {
    handleContextMenuClick(
      {
        menuItemId: "save-to-bookleaf",
        srcUrl: "https://scontent.facebook.com/abc123.jpg",
        pageUrl: "https://www.facebook.com/username/posts/123",
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

describe("handleDragSaveMessage", () => {
  const tab = { id: 1, title: "Tab Title", url: "https://example.com/page" } as browser.Tabs.Tab;

  beforeEach(() => {
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

  it("uses linkUrl as pageUrl when it matches a permalink rule", async () => {
    handleDragSaveMessage(
      {
        type: "drag-save",
        srcUrl: "https://pbs.twimg.com/media/img.jpg",
        linkUrl: "https://x.com/username/status/123456789",
      },
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

  it("falls back to the tab's URL as pageUrl when linkUrl matches no permalink rule", async () => {
    handleDragSaveMessage(
      {
        type: "drag-save",
        srcUrl: "https://example.com/img.jpg",
        linkUrl: "https://example.com/unrelated",
      },
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

  it("falls back to the tab's title when no title was captured", async () => {
    handleDragSaveMessage({ type: "drag-save", srcUrl: "https://example.com/img.jpg" }, tab);
    await vi.waitFor(() => expect(apiFetch).toHaveBeenCalled());

    expect(apiFetch).toHaveBeenCalledWith(
      "/images",
      expect.objectContaining({
        body: expect.stringContaining('"title":"Tab Title"'),
      }),
    );
  });

  it("invokes handleSave with the captured srcUrl, title, and tabId", async () => {
    handleDragSaveMessage(
      { type: "drag-save", srcUrl: "https://example.com/img.jpg", title: "Captured Title" },
      tab,
    );
    await vi.waitFor(() => expect(apiFetch).toHaveBeenCalled());

    expect(fetch).toHaveBeenCalledWith("https://example.com/img.jpg");
    expect(apiFetch).toHaveBeenCalledWith(
      "/images",
      expect.objectContaining({
        body: expect.stringContaining('"title":"Captured Title"'),
      }),
    );
  });
});

import { describe, expect, it } from "vitest";
import { parseSrcset, resolveBestImageSrc, collectPageImages } from "./imagePicker";

function mockImg(
  overrides: Partial<{
    src: string;
    srcset: string;
    dataSrcset: string;
    naturalWidth: number;
    naturalHeight: number;
    complete: boolean;
  }> = {},
): HTMLImageElement {
  const { dataSrcset, ...rest } = overrides;
  const attrs: Record<string, string> = {};
  if (dataSrcset !== undefined) attrs["data-srcset"] = dataSrcset;
  return {
    src: "",
    srcset: "",
    naturalWidth: 100,
    naturalHeight: 100,
    complete: true,
    ...rest,
    getAttribute: (name: string) => attrs[name] ?? null,
  } as unknown as HTMLImageElement;
}

describe("parseSrcset", () => {
  it("returns the widest URL and width from multiple w-descriptor entries", () => {
    expect(parseSrcset("small.jpg 400w, medium.jpg 800w, large.jpg 2400w")).toEqual({
      url: "large.jpg",
      width: 2400,
    });
  });

  it("returns the URL and width from a single w-descriptor entry", () => {
    expect(parseSrcset("image.jpg 1200w")).toEqual({ url: "image.jpg", width: 1200 });
  });

  it("returns null when srcset has no w-descriptors", () => {
    expect(parseSrcset("image.jpg 2x, image-small.jpg 1x")).toBeNull();
  });

  it("returns null for an empty string", () => {
    expect(parseSrcset("")).toBeNull();
  });
});

describe("resolveBestImageSrc", () => {
  it("returns the widest srcset URL and its width when srcset has w-descriptors", () => {
    const img = mockImg({ src: "small.jpg", srcset: "small.jpg 400w, large.jpg 1600w" });
    expect(resolveBestImageSrc(img)).toEqual({ src: "large.jpg", srcsetWidth: 1600 });
  });

  it("falls back to data-srcset when srcset is empty", () => {
    const img = mockImg({
      src: "placeholder.gif",
      srcset: "",
      dataSrcset: "small.png 550w, large.png 1024w",
    });
    expect(resolveBestImageSrc(img)).toEqual({ src: "large.png", srcsetWidth: 1024 });
  });

  it("falls through to high-res platform rule when no srcset", () => {
    const img = mockImg({
      src: "https://pbs.twimg.com/media/ABC?format=jpg&name=small",
      srcset: "",
    });
    expect(resolveBestImageSrc(img)).toEqual({
      src: "https://pbs.twimg.com/media/ABC?format=jpg&name=orig",
      srcsetWidth: null,
    });
  });

  it("falls through to src when no srcset and no platform rule matches", () => {
    const img = mockImg({ src: "https://example.com/photo.jpg", srcset: "" });
    expect(resolveBestImageSrc(img)).toEqual({
      src: "https://example.com/photo.jpg",
      srcsetWidth: null,
    });
  });
});

describe("collectPageImages", () => {
  function makeDoc(imgs: ReturnType<typeof mockImg>[]): Document {
    return {
      querySelectorAll: () => imgs,
    } as unknown as Document;
  }

  it("sorts by srcsetWidth when available, falling back to naturalWidth", () => {
    // naturalWidth says 550, but srcset says 1024 — should sort above a 600px naturalWidth image
    const srcsetLarge = mockImg({
      src: "original.png",
      srcset: "thumb.png 550w, original.png 1024w",
      naturalWidth: 550,
      naturalHeight: 550,
    });
    const naturalLarge = mockImg({ src: "big.jpg", naturalWidth: 600, naturalHeight: 400 });
    const small = mockImg({ src: "small.jpg", naturalWidth: 100, naturalHeight: 100 });

    const result = collectPageImages(makeDoc([naturalLarge, small, srcsetLarge]));

    expect(result.map((i) => i.src)).toEqual(["original.png", "big.jpg", "small.jpg"]);
  });

  it("exposes srcsetWidth on PageImage", () => {
    const img = mockImg({
      src: "original.png",
      srcset: "thumb.png 550w, original.png 1024w",
      naturalWidth: 550,
      naturalHeight: 550,
    });

    const result = collectPageImages(makeDoc([img]));

    expect(result[0].srcsetWidth).toBe(1024);
  });

  it("excludes lazy-loaded images not yet decoded", () => {
    const lazy = mockImg({ src: "lazy.jpg", complete: false, naturalWidth: 0 });
    const loaded = mockImg({ src: "loaded.jpg", complete: true, naturalWidth: 800 });

    const result = collectPageImages(makeDoc([lazy, loaded]));

    expect(result).toHaveLength(1);
    expect(result[0].src).toBe("loaded.jpg");
  });

  it("excludes images with empty src", () => {
    const empty = mockImg({ src: "", srcset: "" });
    const valid = mockImg({ src: "https://example.com/img.jpg" });

    const result = collectPageImages(makeDoc([empty, valid]));

    expect(result).toHaveLength(1);
    expect(result[0].src).toBe("https://example.com/img.jpg");
  });

  it("excludes images with blob src", () => {
    const blob = mockImg({ src: "blob:https://example.com/abc-123", srcset: "" });
    const valid = mockImg({ src: "https://example.com/img.jpg" });

    const result = collectPageImages(makeDoc([blob, valid]));

    expect(result).toHaveLength(1);
    expect(result[0].src).toBe("https://example.com/img.jpg");
  });

  it("deduplicates images with the same resolved src", () => {
    const a = mockImg({ src: "https://example.com/img.jpg", naturalWidth: 800 });
    const b = mockImg({ src: "https://example.com/img.jpg", naturalWidth: 800 });

    const result = collectPageImages(makeDoc([a, b]));

    expect(result).toHaveLength(1);
  });

  it("returns empty array when all images are excluded", () => {
    const lazy = mockImg({ src: "lazy.jpg", complete: false, naturalWidth: 0 });

    expect(collectPageImages(makeDoc([lazy]))).toEqual([]);
  });
});

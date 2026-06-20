import { describe, expect, it } from "vitest";
import { resolveDragImageSrc } from "./dragImageResolveRule";

function mockElement(
  tagName: string,
  options: {
    src?: string;
    closestLink?: { querySelector: (selector: string) => unknown } | null;
    descendantImg?: { src: string } | null;
  } = {},
): Element {
  return {
    tagName,
    src: options.src,
    closest: () => options.closestLink ?? null,
    querySelector: () => options.descendantImg ?? null,
  } as unknown as Element;
}

describe("resolveDragImageSrc", () => {
  it("delegates to resolveCardImageSrc on a registered card site", () => {
    const img = { src: "https://i.pinimg.com/736x/aa/bb/cc.jpg" };
    const link = { querySelector: () => img };
    const target = mockElement("A", { closestLink: link });

    expect(resolveDragImageSrc(target, "https://www.pinterest.com/pin/123")).toBe(img.src);
  });

  it("returns target.src for a plain img on a non-card site", () => {
    const target = mockElement("IMG", { src: "https://example.com/img.jpg" });

    expect(resolveDragImageSrc(target, "https://example.com/page")).toBe("https://example.com/img.jpg");
  });

  it("falls back to the generic img check on a card site when resolveCardImageSrc finds nothing", () => {
    const target = mockElement("IMG", { src: "https://i.pinimg.com/originals/full.jpg" });

    expect(resolveDragImageSrc(target, "https://www.pinterest.com/pin/123")).toBe(
      "https://i.pinimg.com/originals/full.jpg",
    );
  });

  it("falls back to a descendant img when target is a non-img drag wrapper", () => {
    const target = mockElement("DIV", { descendantImg: { src: "https://pbs.twimg.com/media/img.jpg" } });

    expect(resolveDragImageSrc(target, "https://x.com/username")).toBe(
      "https://pbs.twimg.com/media/img.jpg",
    );
  });

  it("returns null for a non-card, non-img target with no descendant img", () => {
    const target = mockElement("DIV");

    expect(resolveDragImageSrc(target, "https://example.com/page")).toBeNull();
  });
});

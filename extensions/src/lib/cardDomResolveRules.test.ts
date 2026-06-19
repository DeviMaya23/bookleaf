import { describe, expect, it } from "vitest";
import { resolveCardImageSrc, shouldResolveCardDom } from "./cardDomResolveRules";

function mockElement(closestLink: { querySelector: (selector: string) => unknown } | null): Element {
  return {
    closest: () => closestLink,
  } as unknown as Element;
}

describe("resolveCardImageSrc", () => {
  it("finds an image when the enclosing link contains an img descendant", () => {
    const img = { src: "https://i.pinimg.com/736x/aa/bb/cc.jpg" };
    const link = { querySelector: () => img };
    const target = mockElement(link);

    expect(resolveCardImageSrc(target)).toBe(img.src);
  });

  it("returns null when there is no enclosing link", () => {
    const target = mockElement(null);

    expect(resolveCardImageSrc(target)).toBeNull();
  });

  it("returns null when the enclosing link has no descendant img", () => {
    const link = { querySelector: () => null };
    const target = mockElement(link);

    expect(resolveCardImageSrc(target)).toBeNull();
  });
});

describe("shouldResolveCardDom", () => {
  it("matches a Pinterest page URL", () => {
    expect(shouldResolveCardDom("https://www.pinterest.com/pin/123")).toBe(true);
  });

  it("matches a Pinterest locale subdomain", () => {
    expect(shouldResolveCardDom("https://id.pinterest.com/pin/123")).toBe(true);
  });

  it("does not match an unregistered site", () => {
    expect(shouldResolveCardDom("https://example.com")).toBe(false);
  });
});

import { describe, expect, it } from "vitest";
import { isImgurUrl, isInstagramUrl, resolveAltText } from "./altTextResolveRule";

function mockImgElement(alt: string | null): Element {
  return {
    tagName: "IMG",
    alt,
  } as unknown as Element;
}

function mockNonImgElement(closestLink: Element | null = null): Element {
  return {
    tagName: "DIV",
    closest: (selector: string) => (selector === "a" ? closestLink : null),
  } as unknown as Element;
}

function mockLinkElement(img: Element | null): Element {
  return {
    tagName: "A",
    querySelector: (selector: string) => (selector === "img" ? img : null),
  } as unknown as Element;
}

describe("isImgurUrl", () => {
  it("matches an imgur.com URL", () => {
    expect(isImgurUrl("https://imgur.com/gallery/abc123")).toBe(true);
  });

  it("does not match an unregistered site", () => {
    expect(isImgurUrl("https://example.com")).toBe(false);
  });
});

describe("isInstagramUrl", () => {
  it("matches an instagram.com URL", () => {
    expect(isInstagramUrl("https://www.instagram.com/p/abc123")).toBe(true);
  });

  it("does not match an unregistered site", () => {
    expect(isInstagramUrl("https://example.com")).toBe(false);
  });
});

describe("resolveAltText", () => {
  it("returns the alt string when present", () => {
    expect(resolveAltText(mockImgElement("a cute cat"))).toBe("a cute cat");
  });

  it("returns null when alt is empty", () => {
    expect(resolveAltText(mockImgElement(""))).toBeNull();
  });

  it("returns null when alt is absent", () => {
    expect(resolveAltText(mockImgElement(null))).toBeNull();
  });

  it("returns null when target is not an <img> and has no enclosing link", () => {
    expect(resolveAltText(mockNonImgElement())).toBeNull();
  });

  it("finds the nested <img> via the closest enclosing link when target wraps it (e.g. a grid card)", () => {
    const img = mockImgElement("a cute cat");
    const link = mockLinkElement(img);
    const wrapper = mockNonImgElement(link);

    expect(resolveAltText(wrapper)).toBe("a cute cat");
  });

  it("returns null when the enclosing link has no nested <img>", () => {
    const link = mockLinkElement(null);
    const wrapper = mockNonImgElement(link);

    expect(resolveAltText(wrapper)).toBeNull();
  });
});

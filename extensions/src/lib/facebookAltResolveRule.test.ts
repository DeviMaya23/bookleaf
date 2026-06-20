import { describe, expect, it } from "vitest";
import { MAX_CLIMB_DEPTH } from "./tweetTextResolveRule";
import { isFacebookUrl, resolveFacebookAltText } from "./facebookAltResolveRule";

function mockElement(
  options: { alt?: string; ariaLabel?: string | null; parent?: Element | null } = {},
): Element {
  const { alt, ariaLabel = null, parent = null } = options;
  return {
    tagName: "IMG",
    alt,
    parentElement: parent,
    closest: () => null,
    getAttribute: (name: string) => (name === "aria-label" ? ariaLabel : null),
  } as unknown as Element;
}

function mockWrapperElement(
  img: Element | null,
  options: { ariaLabel?: string | null; parent?: Element | null } = {},
): Element {
  const { ariaLabel = null, parent = null } = options;
  return {
    tagName: "DIV",
    parentElement: parent,
    closest: (selector: string) =>
      selector === "a" ? ({ querySelector: (s: string) => (s === "img" ? img : null) } as unknown as Element) : null,
    getAttribute: (name: string) => (name === "aria-label" ? ariaLabel : null),
  } as unknown as Element;
}

describe("isFacebookUrl", () => {
  it("matches a facebook.com URL", () => {
    expect(isFacebookUrl("https://www.facebook.com/username/posts/123")).toBe(true);
  });

  it("does not match an unregistered site", () => {
    expect(isFacebookUrl("https://example.com")).toBe(false);
  });
});

describe("resolveFacebookAltText", () => {
  it("resolves from the image's own alt when present", () => {
    const target = mockElement({ alt: "May be an image of gelato" });

    expect(resolveFacebookAltText(target)).toBe("May be an image of gelato");
  });

  it("falls back to an ancestor's aria-label when alt is empty", () => {
    const ancestor = mockElement({ ariaLabel: "May be an image of gelato and text" });
    const target = mockElement({ alt: "", parent: ancestor });

    expect(resolveFacebookAltText(target)).toBe("May be an image of gelato and text");
  });

  it("returns null when neither alt nor an ancestor aria-label is found within the climb bound", () => {
    let current: Element | null = null;
    for (let i = 0; i < MAX_CLIMB_DEPTH + 2; i += 1) {
      current = mockElement({ parent: current });
    }
    const target = mockElement({ alt: "", parent: current });

    expect(resolveFacebookAltText(target)).toBeNull();
  });

  it("finds the nested <img>'s alt via the closest enclosing link when target wraps it", () => {
    const img = mockElement({ alt: "May be an image of gelato" });
    const wrapper = mockWrapperElement(img);

    expect(resolveFacebookAltText(wrapper)).toBe("May be an image of gelato");
  });

  it("falls back to an ancestor's aria-label when the wrapped target has no nested img alt", () => {
    const ancestor = mockElement({ ariaLabel: "May be an image of gelato and text" });
    const wrapper = mockWrapperElement(null, { parent: ancestor });

    expect(resolveFacebookAltText(wrapper)).toBe("May be an image of gelato and text");
  });
});

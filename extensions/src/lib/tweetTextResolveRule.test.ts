import { describe, expect, it } from "vitest";
import { MAX_CLIMB_DEPTH, isTwitterUrl, resolveTweetText } from "./tweetTextResolveRule";

function mockElement(
  textMatch: { textContent: string | null } | null,
  parent: Element | null = null,
): Element {
  return {
    querySelector: () => textMatch,
    parentElement: parent,
  } as unknown as Element;
}

describe("resolveTweetText", () => {
  it("resolves tweetText found directly from the click target's ancestor", () => {
    const tweetText = { textContent: "hello world" };
    const target = mockElement(tweetText);

    expect(resolveTweetText(target)).toBe("hello world");
  });

  it("resolves the nearest tweetText when nested ancestors both contain a match", () => {
    const outerTweetText = { textContent: "outer quoting tweet" };
    const outer = mockElement(outerTweetText);
    const innerTweetText = { textContent: "inner quoted tweet" };
    const inner = mockElement(innerTweetText, outer);
    const target = mockElement(null, inner);

    expect(resolveTweetText(target)).toBe("inner quoted tweet");
  });

  it("returns null when no ancestor within the climb bound contains tweetText", () => {
    let current: Element | null = null;
    for (let i = 0; i < MAX_CLIMB_DEPTH + 2; i += 1) {
      current = mockElement(null, current);
    }

    expect(resolveTweetText(current as Element)).toBeNull();
  });

  it("does not find a match beyond the climb-depth bound", () => {
    const farTweetText = { textContent: "too far up" };
    let current: Element | null = mockElement(farTweetText, null);
    for (let i = 0; i < MAX_CLIMB_DEPTH + 1; i += 1) {
      current = mockElement(null, current);
    }
    const target = current as Element;

    expect(resolveTweetText(target)).toBeNull();
  });
});

describe("isTwitterUrl", () => {
  it("matches an x.com URL", () => {
    expect(isTwitterUrl("https://x.com/username/status/123")).toBe(true);
  });

  it("matches a twitter.com URL", () => {
    expect(isTwitterUrl("https://twitter.com/username")).toBe(true);
  });

  it("does not match an unregistered site", () => {
    expect(isTwitterUrl("https://example.com")).toBe(false);
  });
});

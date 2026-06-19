import { describe, expect, it } from "vitest";
import { extractTwitterHandle, resolveLinkPermalink } from "./linkPermalinkRules";

describe("resolveLinkPermalink", () => {
  it("matches a Twitter status permalink", () => {
    expect(resolveLinkPermalink("https://x.com/username/status/123456789/photo/1")).toBe(true);
  });

  it("matches a Facebook post permalink", () => {
    expect(resolveLinkPermalink("https://www.facebook.com/someuser/posts/123456789")).toBe(true);
  });

  it("does not match a URL from an unregistered site", () => {
    expect(resolveLinkPermalink("https://example.com/some/page")).toBe(false);
  });
});

describe("extractTwitterHandle", () => {
  it("extracts the handle from a Twitter status permalink", () => {
    expect(extractTwitterHandle("https://x.com/username/status/123456789/photo/1")).toBe(
      "username",
    );
  });

  it("returns null for a non-Twitter URL", () => {
    expect(extractTwitterHandle("https://www.facebook.com/someuser/posts/123456789")).toBeNull();
  });

  it("returns null for a Twitter URL with no status path", () => {
    expect(extractTwitterHandle("https://x.com/username")).toBeNull();
  });
});

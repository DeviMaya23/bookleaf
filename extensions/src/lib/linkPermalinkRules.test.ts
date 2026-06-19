import { describe, expect, it } from "vitest";
import { resolveLinkPermalink } from "./linkPermalinkRules";

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

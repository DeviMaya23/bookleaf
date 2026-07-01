import { describe, expect, it } from "vitest";
import { cleanUrl, type CompiledProvider } from "./urlCleaner";

describe("cleanUrl", () => {
  it("strips known tracking params from a Google search URL", () => {
    const result = cleanUrl(
      "https://www.google.com/search?q=cats&tbm=isch&ei=XyZ&ved=AbC&biw=1234&source=hp",
    );
    expect(result).toBe("https://www.google.com/search?q=cats&tbm=isch");
  });

  it("strips a referralMarketing param (Google referrer) from a Google URL", () => {
    const result = cleanUrl(
      "https://www.google.com/search?q=cats&referrer=homepage",
    );
    expect(result).toBe("https://www.google.com/search?q=cats");
  });

  it("strips Twitter tracking params (s, t, ref_src)", () => {
    const result = cleanUrl(
      "https://twitter.com/user/status/123456?s=20&t=abc123&ref_src=twsrc%5Etfw",
    );
    expect(result).toBe("https://twitter.com/user/status/123456");
  });

  it("does not clean a Twitter redirect URL (exception vetoes rule application)", () => {
    const input = "https://twitter.com/i/redirect?url=https%3A%2F%2Fexample.com&s=20";
    expect(cleanUrl(input)).toBe(input);
  });

  it("strips Facebook ref param (matches [a-z]*ref[a-z]* rule)", () => {
    const result = cleanUrl(
      "https://www.facebook.com/photo?fbid=123456789&ref=share&__tn__=EH-R",
    );
    expect(result).toBe("https://www.facebook.com/photo?fbid=123456789");
  });

  it("strips Reddit tracking params", () => {
    const result = cleanUrl(
      "https://www.reddit.com/r/cats/comments/abc123/?ref_campaign=email&ref_source=digest",
    );
    expect(result).toBe("https://www.reddit.com/r/cats/comments/abc123/");
  });

  it("returns an unrecognised host URL unchanged", () => {
    const input = "https://example.com/page?ref=123&tracking=abc";
    expect(cleanUrl(input)).toBe(input);
  });

  it("returns a malformed URL unchanged", () => {
    const input = "not a url";
    expect(cleanUrl(input)).toBe(input);
  });

  it("completeProvider strips all query params and returns immediately", () => {
    const provider: CompiledProvider = {
      urlPattern: /^https?:\/\/example\.com/,
      completeProvider: true,
      rules: [],
      referralMarketing: [],
      exceptions: [],
    };
    const result = cleanUrl("https://example.com/page?a=1&b=2&c=3", [provider]);
    expect(result).toBe("https://example.com/page");
  });

  it("applies rules from multiple matching providers", () => {
    const providerA: CompiledProvider = {
      urlPattern: /^https?:\/\/example\.com/,
      completeProvider: false,
      rules: [/^(?:utm_source)$/],
      referralMarketing: [],
      exceptions: [],
    };
    const providerB: CompiledProvider = {
      urlPattern: /^https?:\/\/example\.com/,
      completeProvider: false,
      rules: [/^(?:fbclid)$/],
      referralMarketing: [],
      exceptions: [],
    };
    const result = cleanUrl("https://example.com/page?q=cats&utm_source=email&fbclid=abc", [providerA, providerB]);
    expect(result).toBe("https://example.com/page?q=cats");
  });
});

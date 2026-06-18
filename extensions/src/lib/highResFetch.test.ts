import { describe, expect, it } from "vitest";
import { resolveHighResUrl, validateDimension, validateResponseShape } from "./highResFetch";

describe("resolveHighResUrl", () => {
  it("returns the transformed URL when a rule matches", () => {
    expect(resolveHighResUrl("https://pbs.twimg.com/media/XXXXX?format=jpg&name=small")).toBe(
      "https://pbs.twimg.com/media/XXXXX?format=jpg&name=orig",
    );
  });

  it("returns null when no rule matches", () => {
    expect(resolveHighResUrl("https://example.com/image.jpg")).toBeNull();
  });

  it("returns null for an invalid URL", () => {
    expect(resolveHighResUrl("not a url")).toBeNull();
  });
});

function makeResponse(ok: boolean, contentType: string | null): Response {
  return {
    ok,
    headers: { get: (name: string) => (name === "Content-Type" ? contentType : null) },
  } as unknown as Response;
}

describe("validateResponseShape", () => {
  it("returns false for a non-OK response", () => {
    expect(validateResponseShape(makeResponse(false, "image/jpeg"))).toBe(false);
  });

  it("returns false for a disallowed content type", () => {
    expect(validateResponseShape(makeResponse(true, "text/html"))).toBe(false);
  });

  it("returns true for an allowed content type", () => {
    expect(validateResponseShape(makeResponse(true, "image/jpeg"))).toBe(true);
  });
});

describe("validateDimension", () => {
  it("returns false when below the threshold", () => {
    expect(validateDimension({ width: 50, height: 200 })).toBe(false);
  });

  it("returns true when at the threshold", () => {
    expect(validateDimension({ width: 100, height: 100 })).toBe(true);
  });

  it("returns true when above the threshold", () => {
    expect(validateDimension({ width: 200, height: 200 })).toBe(true);
  });
});

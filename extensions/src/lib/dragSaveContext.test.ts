import { describe, expect, it } from "vitest";
import { resolveDragSaveContext } from "./dragSaveContext";

function mockImgTarget(src: string): Element {
  return {
    tagName: "IMG",
    src,
    closest: () => null,
    querySelector: () => null,
  } as unknown as Element;
}

describe("resolveDragSaveContext", () => {
  it("does not render a drop zone when drag-to-save is disabled, even with a resolvable srcUrl", () => {
    const target = mockImgTarget("https://example.com/img.jpg");

    expect(resolveDragSaveContext(target, "https://example.com/page", false)).toBeNull();
  });

  it("resolves a context when drag-to-save is enabled and srcUrl resolves", () => {
    const target = mockImgTarget("https://example.com/img.jpg");

    const context = resolveDragSaveContext(target, "https://example.com/page", true);

    expect(context).toEqual({
      srcUrl: "https://example.com/img.jpg",
      title: undefined,
      linkUrl: undefined,
    });
  });

  it("returns null when drag-to-save is enabled but srcUrl does not resolve", () => {
    const target = {
      tagName: "DIV",
      closest: () => null,
      querySelector: () => null,
    } as unknown as Element;

    expect(resolveDragSaveContext(target, "https://example.com/page", true)).toBeNull();
  });
});

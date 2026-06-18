import { describe, expect, it } from "vitest";
import { rules } from "./highResRules";

const twitterRule = rules.find((r) => r.id === "twitter-media")!;
const pinterestRule = rules.find((r) => r.id === "pinterest-size-segment")!;

describe("twitter-media rule", () => {
  it("matches a media image with a resizable name param", () => {
    const url = new URL("https://pbs.twimg.com/media/XXXXX?format=jpg&name=small");
    expect(twitterRule.matches(url)).toBe(true);
  });

  it("does not match a profile image", () => {
    const url = new URL("https://pbs.twimg.com/profile_images/XXXXX/avatar_normal.jpg");
    expect(twitterRule.matches(url)).toBe(false);
  });

  it("transforms name to orig", () => {
    const url = new URL("https://pbs.twimg.com/media/XXXXX?format=jpg&name=small");
    expect(twitterRule.transform(url)).toBe("https://pbs.twimg.com/media/XXXXX?format=jpg&name=orig");
  });
});

describe("pinterest-size-segment rule", () => {
  it("matches a sized image path", () => {
    const url = new URL("https://i.pinimg.com/736x/aa/bb/cc/aabbccdd.jpg");
    expect(pinterestRule.matches(url)).toBe(true);
  });

  it("transforms the size segment to originals", () => {
    const url = new URL("https://i.pinimg.com/736x/aa/bb/cc/aabbccdd.jpg");
    expect(pinterestRule.transform(url)).toBe("https://i.pinimg.com/originals/aa/bb/cc/aabbccdd.jpg");
  });
});

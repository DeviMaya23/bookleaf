import { describe, expect, it } from "vitest";
import { rules } from "./highResRules";

const twitterRule = rules.find((r) => r.id === "twitter-media")!;
const pinterestRule = rules.find((r) => r.id === "pinterest-size-segment")!;
const imgurRule = rules.find((r) => r.id === "imgur-thumbnail")!;
const facebookRule = rules.find((r) => r.id === "facebook-ctp")!;

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

describe("imgur-thumbnail rule", () => {
  it("matches a thumbnail URL with a size suffix", () => {
    const url = new URL("https://i.imgur.com/xCbCj7a_d.webp?maxwidth=520&shape=thumb&fidelity=high");
    expect(imgurRule.matches(url)).toBe(true);
  });

  it("does not match a bare id with no size suffix", () => {
    const url = new URL("https://i.imgur.com/xCbCj7a.jpg");
    expect(imgurRule.matches(url)).toBe(false);
  });

  it("does not match a video delivery URL", () => {
    const mp4 = new URL("https://i.imgur.com/xCbCj7a.mp4");
    const gifv = new URL("https://i.imgur.com/xCbCj7a.gifv");
    expect(imgurRule.matches(mp4)).toBe(false);
    expect(imgurRule.matches(gifv)).toBe(false);
  });

  it("transforms the thumbnail URL to the bare id with a jpg extension", () => {
    const url = new URL("https://i.imgur.com/xCbCj7a_d.webp?maxwidth=520&shape=thumb&fidelity=high");
    expect(imgurRule.transform(url)).toBe("https://i.imgur.com/xCbCj7a.jpg");
  });
});

describe("facebook-ctp rule", () => {
  it("matches an fbcdn.net URL containing ctp", () => {
    const url = new URL(
      "https://scontent-sjc3-1.xx.fbcdn.net/v/t1.0/img.jpg?ctp=s590x590&cstp=mx1631x1087&oh=abc&oe=def",
    );
    expect(facebookRule.matches(url)).toBe(true);
  });

  it("does not match an fbcdn.net URL without ctp", () => {
    const url = new URL(
      "https://scontent-sjc3-1.xx.fbcdn.net/v/t1.0/img.jpg?cstp=mx1631x1087&oh=abc&oe=def",
    );
    expect(facebookRule.matches(url)).toBe(false);
  });

  it("removes only ctp and leaves other params byte-for-byte unchanged", () => {
    const url = new URL(
      "https://scontent-sjc3-1.xx.fbcdn.net/v/t1.0/img.jpg?ctp=s590x590&cstp=mx1631x1087&oh=abc&oe=def&_nc_ohc=xyz",
    );
    expect(facebookRule.transform(url)).toBe(
      "https://scontent-sjc3-1.xx.fbcdn.net/v/t1.0/img.jpg?cstp=mx1631x1087&oh=abc&oe=def&_nc_ohc=xyz",
    );
  });
});

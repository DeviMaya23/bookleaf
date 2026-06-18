export type HighResRule = {
  id: string;
  matches: (url: URL) => boolean;
  transform: (url: URL) => string | null;
};

const TWITTER_MEDIA_PATH = /\/media\//;
const TWITTER_ALLOWED_FORMATS = new Set(["jpg", "png", "webp"]);

function isTwitterMediaImage(url: URL): boolean {
  if (url.hostname !== "pbs.twimg.com") return false;
  if (!TWITTER_MEDIA_PATH.test(url.pathname)) return false;
  const format = url.searchParams.get("format");
  const name = url.searchParams.get("name");
  if (!format || !TWITTER_ALLOWED_FORMATS.has(format.toLowerCase())) return false;
  if (!name || name.toLowerCase() === "orig") return false;
  return true;
}

const twitterRule: HighResRule = {
  id: "twitter-media",
  matches: isTwitterMediaImage,
  transform: (url) => {
    if (!isTwitterMediaImage(url)) return null;
    const newUrl = new URL(url.toString());
    newUrl.searchParams.set("name", "orig");
    return newUrl.toString();
  },
};

const PINTEREST_SIZE_SEGMENT = /^\/(\d+x)\//;

const pinterestRule: HighResRule = {
  id: "pinterest-size-segment",
  matches: (url) =>
    url.hostname === "i.pinimg.com" && PINTEREST_SIZE_SEGMENT.test(url.pathname),
  transform: (url) => {
    const match = url.pathname.match(PINTEREST_SIZE_SEGMENT);
    if (!match) return null;
    const newUrl = new URL(url.toString());
    newUrl.pathname = newUrl.pathname.replace(PINTEREST_SIZE_SEGMENT, "/originals/");
    return newUrl.toString();
  },
};

export const rules: HighResRule[] = [twitterRule, pinterestRule];

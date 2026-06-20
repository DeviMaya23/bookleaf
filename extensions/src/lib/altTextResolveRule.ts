const IMGUR_HOSTNAMES = new Set(["imgur.com", "www.imgur.com", "i.imgur.com"]);
const INSTAGRAM_HOSTNAMES = new Set(["instagram.com", "www.instagram.com"]);

export function isImgurUrl(pageUrl: string): boolean {
  let url: URL;
  try {
    url = new URL(pageUrl);
  } catch {
    return false;
  }
  return IMGUR_HOSTNAMES.has(url.hostname);
}

export function isInstagramUrl(pageUrl: string): boolean {
  let url: URL;
  try {
    url = new URL(pageUrl);
  } catch {
    return false;
  }
  return INSTAGRAM_HOSTNAMES.has(url.hostname);
}

export function findImg(target: Element): HTMLImageElement | null {
  if (target.tagName === "IMG") return target as HTMLImageElement;
  const link = target.closest("a");
  return link?.querySelector("img") ?? null;
}

export function resolveAltText(target: Element): string | null {
  const img = findImg(target);
  return img?.alt || null;
}

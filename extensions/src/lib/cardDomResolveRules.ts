export type CardDomResolveRule = {
  id: string;
  matches: (url: URL) => boolean;
};

const pinterestRule: CardDomResolveRule = {
  id: "pinterest-card",
  matches: (url) => url.hostname === "pinterest.com" || url.hostname.endsWith(".pinterest.com"),
};

export const rules: CardDomResolveRule[] = [pinterestRule];

export function shouldResolveCardDom(pageUrl: string): boolean {
  let url: URL;
  try {
    url = new URL(pageUrl);
  } catch {
    return false;
  }
  return rules.some((rule) => rule.matches(url));
}

export const linkOnlyCardUrlPatterns: string[] = ["*://*.pinterest.com/pin/*"];

export function resolveCardImageSrc(target: Element): string | null {
  const link = target.closest("a");
  if (!link) return null;
  const img = link.querySelector("img");
  if (!img) return null;
  return img.src || null;
}

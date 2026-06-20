import { MAX_CLIMB_DEPTH } from "./tweetTextResolveRule";
import { findImg } from "./altTextResolveRule";

const FACEBOOK_HOSTNAMES = new Set(["facebook.com", "www.facebook.com", "m.facebook.com"]);

export function isFacebookUrl(pageUrl: string): boolean {
  let url: URL;
  try {
    url = new URL(pageUrl);
  } catch {
    return false;
  }
  return FACEBOOK_HOSTNAMES.has(url.hostname);
}

export function resolveFacebookAltText(target: Element): string | null {
  const alt = findImg(target)?.alt;
  if (alt) return alt;

  let current: Element | null = target.parentElement;
  for (let depth = 0; depth < MAX_CLIMB_DEPTH && current; depth += 1) {
    const ariaLabel = current.getAttribute("aria-label");
    if (ariaLabel) return ariaLabel;
    current = current.parentElement;
  }
  return null;
}

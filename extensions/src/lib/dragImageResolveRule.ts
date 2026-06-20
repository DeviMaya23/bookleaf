import { resolveCardImageSrc, shouldResolveCardDom } from "./cardDomResolveRules";

export function resolveDragImageSrc(target: Element, pageUrl: string): string | null {
  if (shouldResolveCardDom(pageUrl)) {
    const cardSrc = resolveCardImageSrc(target);
    if (cardSrc) return cardSrc;
  }
  if (target.tagName === "IMG") return (target as HTMLImageElement).src || null;
  const descendantImg = target.querySelector("img");
  if (descendantImg) return descendantImg.src || null;
  return null;
}

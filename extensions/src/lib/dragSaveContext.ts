import { resolveDragImageSrc } from "./dragImageResolveRule";
import { isTwitterUrl, resolveTweetText } from "./tweetTextResolveRule";
import { isImgurUrl, isInstagramUrl, resolveAltText } from "./altTextResolveRule";
import { isFacebookUrl, resolveFacebookAltText } from "./facebookAltResolveRule";

export interface DragSaveContext {
  srcUrl: string;
  title?: string;
  linkUrl?: string;
}

function resolveDragTitle(target: Element, pageUrl: string): string | undefined {
  if (isTwitterUrl(pageUrl)) return resolveTweetText(target) ?? undefined;
  if (isImgurUrl(pageUrl) || isInstagramUrl(pageUrl)) return resolveAltText(target) ?? undefined;
  if (isFacebookUrl(pageUrl)) return resolveFacebookAltText(target) ?? undefined;
  return undefined;
}

export function resolveDragSaveContext(
  target: Element,
  pageUrl: string,
  dragEnabled: boolean,
): DragSaveContext | null {
  if (!dragEnabled) return null;

  const srcUrl = resolveDragImageSrc(target, pageUrl);
  if (!srcUrl) return null;

  return {
    srcUrl,
    title: resolveDragTitle(target, pageUrl),
    linkUrl: target.closest("a")?.href,
  };
}

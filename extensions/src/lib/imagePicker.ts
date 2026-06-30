import { resolveHighResUrl } from "./highResFetch";

export type PageImage = {
  src: string;
  naturalWidth: number;
  naturalHeight: number;
  srcsetWidth: number | null;
};

export function parseSrcset(srcset: string): { url: string; width: number } | null {
  if (!srcset) return null;

  let best: { url: string; width: number } | null = null;
  for (const part of srcset.split(",")) {
    const tokens = part.trim().split(/\s+/);
    if (tokens.length < 2) continue;
    const descriptor = tokens[tokens.length - 1];
    if (!descriptor.endsWith("w")) continue;
    const w = parseInt(descriptor.slice(0, -1), 10);
    if (isNaN(w)) continue;
    const url = tokens[0];
    if (!best || w > best.width) best = { url, width: w };
  }

  return best ?? null;
}

export function resolveBestImageSrc(
  img: HTMLImageElement,
): { src: string; srcsetWidth: number | null } {
  const srcsetAttr = img.srcset || img.getAttribute("data-srcset") || "";
  if (srcsetAttr) {
    const entry = parseSrcset(srcsetAttr);
    if (entry) return { src: entry.url, srcsetWidth: entry.width };
  }

  const highRes = resolveHighResUrl(img.src);
  if (highRes) return { src: highRes, srcsetWidth: null };

  return { src: img.src, srcsetWidth: null };
}

export function collectPageImages(document: Document): PageImage[] {
  const imgs = Array.from(document.querySelectorAll<HTMLImageElement>("img"));
  const results: PageImage[] = [];

  const seen = new Set<string>();
  for (const img of imgs) {
    if (img.complete === false && img.naturalWidth === 0) continue;

    const { src, srcsetWidth } = resolveBestImageSrc(img);
    if (!src || src.startsWith("blob:") || seen.has(src)) continue;
    seen.add(src);

    results.push({ src, naturalWidth: img.naturalWidth, naturalHeight: img.naturalHeight, srcsetWidth });
  }

  results.sort((a, b) => sortKey(b) - sortKey(a));
  return results;
}

function sortKey(image: PageImage): number {
  if (image.srcsetWidth !== null) return image.srcsetWidth;
  return Math.sqrt(image.naturalWidth * image.naturalHeight);
}

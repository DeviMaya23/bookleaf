import { rules } from "./highResRules";

const ALLOWED_CONTENT_TYPES = new Set(["image/jpeg", "image/png", "image/webp"]);
const MIN_DIMENSION = 100;

export function resolveHighResUrl(srcUrl: string): string | null {
  let url: URL;
  try {
    url = new URL(srcUrl);
  } catch {
    return null;
  }
  for (const rule of rules) {
    if (rule.matches(url)) return rule.transform(url);
  }
  return null;
}

export type CandidateValidation =
  | { valid: true; bitmap: ImageBitmap | null }
  | { valid: false; bitmap: null };

export async function validateCandidate(
  response: Response,
  blob: Blob,
): Promise<CandidateValidation> {
  if (!response.ok) return { valid: false, bitmap: null };

  const contentType = response.headers.get("Content-Type")?.split(";")[0].trim();
  if (!contentType || !ALLOWED_CONTENT_TYPES.has(contentType)) {
    return { valid: false, bitmap: null };
  }

  if (typeof createImageBitmap === "undefined") {
    return { valid: true, bitmap: null };
  }

  const bitmap = await createImageBitmap(blob);
  if (bitmap.width < MIN_DIMENSION || bitmap.height < MIN_DIMENSION) {
    bitmap.close();
    return { valid: false, bitmap: null };
  }

  return { valid: true, bitmap };
}

import browser from "webextension-polyfill";
import { getAuth, addRecentSave, type BookleafAuth } from "../lib/storage";
import { apiFetch } from "../lib/api";
import { resolveHighResReferrer, resolveHighResUrl, validateCandidate } from "../lib/highResFetch";
import { extractTwitterHandle, resolveLinkPermalink } from "../lib/linkPermalinkRules";
import { linkOnlyCardUrlPatterns } from "../lib/cardDomResolveRules";

const isProductionBuild =
  import.meta.env.MODE === "chrome-production" || import.meta.env.MODE === "firefox-production";

if (!isProductionBuild) {
  browser.action.setBadgeText({ text: "DEV" });
  browser.action.setBadgeBackgroundColor({ color: "#d97706" });
}

browser.runtime.onInstalled.addListener(async () => {
  await browser.contextMenus.removeAll();
  browser.contextMenus.create({
    id: "save-to-bookleaf",
    title: "Save to Bookleaf",
    contexts: ["image"],
  });
  browser.contextMenus.create({
    id: "save-to-bookleaf-link",
    title: "Save to Bookleaf",
    contexts: ["link"],
    targetUrlPatterns: linkOnlyCardUrlPatterns,
  });
});

export const resolvedContextByTab = new Map<number, Partial<{ srcUrl: string; title: string }>>();

interface DragSaveMessage {
  type: "drag-save";
  srcUrl: string;
  title?: string;
  linkUrl?: string;
}

export function handleDragSaveMessage(
  msg: DragSaveMessage,
  tab: browser.Tabs.Tab | undefined,
): void {
  const pageUrl = msg.linkUrl && resolveLinkPermalink(msg.linkUrl) ? msg.linkUrl : tab?.url ?? "";
  const title = msg.title ?? tab?.title ?? "Untitled";
  handleSave({ srcUrl: msg.srcUrl, pageUrl, title, tabId: tab?.id });
}

interface SnipCapturedMessage {
  type: "snip-captured";
  blob: Blob;
  mimeType: string;
}

export function handleSnipCapturedMessage(
  msg: SnipCapturedMessage,
  tab: browser.Tabs.Tab | undefined,
): void {
  const pageUrl = tab?.url ?? "";
  const title = tab?.title ?? "Untitled";
  handleCapture({ blob: msg.blob, mimeType: msg.mimeType, pageUrl, title, tabId: tab?.id });
}

browser.runtime.onMessage.addListener((message: unknown, sender: browser.Runtime.MessageSender) => {
  const msg = message as {
    type?: string;
    resolved?: Partial<{ srcUrl: string; title: string }>;
  };

  if (msg.type === "drag-save") {
    handleDragSaveMessage(message as DragSaveMessage, sender.tab);
    return;
  }

  if (msg.type === "snip-captured") {
    handleSnipCapturedMessage(message as SnipCapturedMessage, sender.tab);
    return;
  }

  if (!msg.resolved || sender.tab?.id === undefined) return;
  resolvedContextByTab.set(sender.tab.id, msg.resolved);
});

const SNIP_COMMAND = "snip-capture";

export async function handleSnipCommand(command: string): Promise<void> {
  if (command !== SNIP_COMMAND) return;

  const [tab] = await browser.tabs.query({ active: true, currentWindow: true });
  if (tab?.id === undefined) return;

  try {
    const dataUrl = await browser.tabs.captureVisibleTab(tab.windowId);
    await browser.tabs.sendMessage(tab.id, { type: "snip-frame", dataUrl });
  } catch {
    // content script may not be injected on this page (e.g. chrome://, PDF viewer) — no-op
  }
}

browser.commands.onCommand.addListener(handleSnipCommand);

function resolveTitle(
  info: browser.Menus.OnClickData,
  tab: browser.Tabs.Tab | undefined,
  resolved: Partial<{ srcUrl: string; title: string }> | undefined,
): string {
  const handle = info.linkUrl ? extractTwitterHandle(info.linkUrl) : null;
  if (handle) {
    return resolved?.title ? `@${handle}: ${resolved.title.slice(0, 100)}...` : `@${handle}`;
  }
  if (resolved?.title) return resolved.title;
  return tab?.title ?? info.pageUrl ?? "Untitled";
}

export function handleContextMenuClick(
  info: browser.Menus.OnClickData,
  tab: browser.Tabs.Tab | undefined,
): void {
  const pageUrl = info.pageUrl;
  const resolved = tab?.id !== undefined ? resolvedContextByTab.get(tab.id) : undefined;
  const title = resolveTitle(info, tab, resolved);

  if (info.menuItemId === "save-to-bookleaf") {
    const srcUrl = info.srcUrl;
    if (!srcUrl) return;
    const sourceUrl =
      info.linkUrl && resolveLinkPermalink(info.linkUrl) ? info.linkUrl : pageUrl ?? "";
    handleSave({ srcUrl, pageUrl: sourceUrl, title, tabId: tab?.id });
    return;
  }

  if (info.menuItemId === "save-to-bookleaf-link") {
    if (!resolved?.srcUrl) {
      void sendToast(tab?.id, "error", "Couldn't save image.", "Check your connection and try again.");
      return;
    }
    handleSave({ srcUrl: resolved.srcUrl, pageUrl: info.linkUrl ?? "", title, tabId: tab?.id });
  }
}

browser.contextMenus.onClicked.addListener(handleContextMenuClick);

export function isTokenValid(auth: BookleafAuth | null): auth is BookleafAuth {
  if (!auth) return false;
  return Date.now() < auth.expiresAt;
}

async function fetchImageBlob(
  url: string,
): Promise<{ blob: Blob; mimeType: string }> {
  const response = await fetch(url);
  if (!response.ok) throw new Error(`Failed to fetch image: ${response.status}`);
  const blob = await response.blob();
  const mimeType = blob.type || "image/jpeg";
  return { blob, mimeType };
}

export async function resolveImageBlob(
  srcUrl: string,
): Promise<{ blob: Blob; mimeType: string; bitmap: ImageBitmap | null }> {
  const candidateUrl = resolveHighResUrl(srcUrl);
  if (candidateUrl) {
    try {
      const referrer = resolveHighResReferrer(srcUrl);
      const response = referrer ? await fetch(candidateUrl, { referrer }) : await fetch(candidateUrl);
      const blob = await response.blob();
      const validation = await validateCandidate(response, blob);
      if (validation.valid) {
        return { blob, mimeType: blob.type || "image/jpeg", bitmap: validation.bitmap };
      }
    } catch {
      // candidate fetch/validation failed — fall through to the original srcUrl
    }
  }

  const fallback = await fetchImageBlob(srcUrl);
  return { ...fallback, bitmap: null };
}

async function generateThumbnail(
  blob: Blob,
  existingBitmap?: ImageBitmap | null,
): Promise<{ blob: Blob; width: number; height: number }> {
  const bitmap = existingBitmap ?? (await createImageBitmap(blob));

  const { width, height } = bitmap;
  const scale = Math.min(1, 600 / Math.max(width, height));
  const tw = Math.round(width * scale);
  const th = Math.round(height * scale);

  const canvas = new OffscreenCanvas(tw, th);
  const ctx = canvas.getContext("2d")!;
  ctx.drawImage(bitmap, 0, 0, tw, th);
  bitmap.close();

  const thumbnailBlob = await canvas.convertToBlob({ type: "image/jpeg", quality: 0.9 });
  return { blob: thumbnailBlob, width, height };
}

export async function blobToDataUrl(blob: Blob): Promise<string> {
  const buffer = await blob.arrayBuffer();
  const bytes = new Uint8Array(buffer);
  let binary = "";
  const chunkSize = 8192;
  for (let i = 0; i < bytes.length; i += chunkSize) {
    binary += String.fromCharCode(...bytes.subarray(i, i + chunkSize));
  }
  return `data:image/jpeg;base64,${btoa(binary)}`;
}

export async function saveImage({
  blob,
  mimeType,
  title,
  pageUrl,
  thumbnailBlob,
  width,
  height,
}: {
  blob: Blob;
  mimeType: string;
  title: string;
  pageUrl: string;
  accessToken: string;
  thumbnailBlob?: Blob;
  width?: number;
  height?: number;
}): Promise<string> {
  const initRes = await apiFetch("/images", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      title,
      mime_type: mimeType,
      source_url: pageUrl || undefined,
    }),
  });
  if (!initRes.ok) throw new Error(`POST /images failed: ${initRes.status}`);

  const { upload_url, thumbnail_upload_url, id: image_id } = (await initRes.json()) as {
    upload_url: string;
    thumbnail_upload_url: string;
    id: string;
  };

  const puts: Promise<Response>[] = [
    fetch(upload_url, {
      method: "PUT",
      headers: { "Content-Type": mimeType },
      body: blob,
    }),
  ];
  if (thumbnailBlob) {
    puts.push(
      fetch(thumbnail_upload_url, {
        method: "PUT",
        headers: { "Content-Type": "image/jpeg" },
        body: thumbnailBlob,
      }),
    );
  }

  const results = await Promise.all(puts);
  for (const res of results) {
    if (!res.ok) throw new Error(`PUT to R2 failed: ${res.status}`);
  }

  const completeBody: { file_size: number; width?: number; height?: number } = {
    file_size: blob.size,
  };
  if (width !== undefined) completeBody.width = width;
  if (height !== undefined) completeBody.height = height;

  const completeRes = await apiFetch(`/images/${image_id}/complete`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(completeBody),
  });
  if (!completeRes.ok)
    throw new Error(`POST /complete failed: ${completeRes.status}`);

  return image_id;
}

export async function persistImage({
  blob,
  mimeType,
  bitmap,
  title,
  pageUrl,
  tabId,
}: {
  blob: Blob;
  mimeType: string;
  bitmap?: ImageBitmap | null;
  title: string;
  pageUrl: string;
  tabId: number | undefined;
}): Promise<void> {
  const auth = await getAuth();

  if (!isTokenValid(auth)) {
    await sendToast(tabId, "error", "Bookleaf", "Please log in first.");
    return;
  }

  let imageId: string | null = null;
  let thumbnailBlob: Blob | null = null;

  try {
    let dimensions: { width: number; height: number } | undefined;
    if (typeof OffscreenCanvas !== "undefined") {
      const thumbnail = await generateThumbnail(blob, bitmap);
      thumbnailBlob = thumbnail.blob;
      dimensions = { width: thumbnail.width, height: thumbnail.height };
    } else {
      bitmap?.close();
    }

    imageId = await saveImage({
      blob,
      mimeType,
      title,
      pageUrl,
      accessToken: auth.accessToken,
      thumbnailBlob: thumbnailBlob ?? undefined,
      width: dimensions?.width,
      height: dimensions?.height,
    });
    await sendToast(tabId, "success", "Saved to Bookleaf.", "Added to Unsorted.");
  } catch {
    await sendToast(tabId, "error", "Couldn't save image.", "Check your connection and try again.");
    return;
  }

  if (imageId) {
    const dataUrl = thumbnailBlob ? await blobToDataUrl(thumbnailBlob) : "";
    await addRecentSave({ imageId, title, dataUrl, savedAt: Date.now() });
  }
}

export async function handleSave({
  srcUrl,
  pageUrl,
  title,
  tabId,
}: {
  srcUrl: string;
  pageUrl: string;
  title: string;
  tabId: number | undefined;
}): Promise<void> {
  const auth = await getAuth();
  if (!isTokenValid(auth)) {
    await sendToast(tabId, "error", "Bookleaf", "Please log in first.");
    return;
  }

  let fetched: { blob: Blob; mimeType: string; bitmap: ImageBitmap | null };
  try {
    fetched = await resolveImageBlob(srcUrl);
  } catch {
    await sendToast(tabId, "error", "Couldn't save image.", "Check your connection and try again.");
    return;
  }

  await persistImage({ ...fetched, title, pageUrl, tabId });
}

export async function handleCapture({
  blob,
  mimeType,
  pageUrl,
  title,
  tabId,
}: {
  blob: Blob;
  mimeType: string;
  pageUrl: string;
  title: string;
  tabId: number | undefined;
}): Promise<void> {
  await persistImage({ blob, mimeType, title, pageUrl, tabId });
}

async function sendToast(
  tabId: number | undefined,
  variant: "success" | "error",
  title: string,
  body: string,
): Promise<void> {
  if (tabId === undefined) return;
  try {
    await browser.tabs.sendMessage(tabId, { type: "toast", variant, title, body });
  } catch {
    // tab may have navigated away — silently ignore
  }
}

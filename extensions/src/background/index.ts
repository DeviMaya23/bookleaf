import browser from "webextension-polyfill";
import { getAuth, addRecentSave, type BookleafAuth } from "../lib/storage";
import { apiFetch } from "../lib/api";

browser.runtime.onInstalled.addListener(async () => {
  await browser.contextMenus.removeAll();
  browser.contextMenus.create({
    id: "save-to-bookleaf",
    title: "Save to Bookleaf",
    contexts: ["image"],
  });
});

browser.contextMenus.onClicked.addListener((info, tab) => {
  if (info.menuItemId !== "save-to-bookleaf") return;
  const srcUrl = info.srcUrl;
  const pageUrl = info.pageUrl;
  const title = tab?.title ?? pageUrl ?? "Untitled";
  if (!srcUrl) return;
  handleSave({ srcUrl, pageUrl: pageUrl ?? "", title, tabId: tab?.id });
});

function isTokenValid(auth: BookleafAuth | null): auth is BookleafAuth {
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

async function generateThumbnail(blob: Blob): Promise<Blob> {
  const bitmap = await createImageBitmap(blob);

  const { width, height } = bitmap;
  const scale = Math.min(1, 600 / Math.max(width, height));
  const tw = Math.round(width * scale);
  const th = Math.round(height * scale);

  const canvas = new OffscreenCanvas(tw, th);
  const ctx = canvas.getContext("2d")!;
  ctx.drawImage(bitmap, 0, 0, tw, th);
  bitmap.close();

  return canvas.convertToBlob({ type: "image/jpeg", quality: 0.9 });
}

async function blobToDataUrl(blob: Blob): Promise<string> {
  const buffer = await blob.arrayBuffer();
  const bytes = new Uint8Array(buffer);
  let binary = "";
  const chunkSize = 8192;
  for (let i = 0; i < bytes.length; i += chunkSize) {
    binary += String.fromCharCode(...bytes.subarray(i, i + chunkSize));
  }
  return `data:image/jpeg;base64,${btoa(binary)}`;
}

async function saveImage({
  blob,
  mimeType,
  title,
  pageUrl,
  thumbnailBlob,
}: {
  blob: Blob;
  mimeType: string;
  title: string;
  pageUrl: string;
  accessToken: string;
  thumbnailBlob?: Blob;
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

  const completeRes = await apiFetch(`/images/${image_id}/complete`, {
    method: "POST",
  });
  if (!completeRes.ok)
    throw new Error(`POST /complete failed: ${completeRes.status}`);

  return image_id;
}

async function handleSave({
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

  let imageId: string | null = null;
  let thumbnailBlob: Blob | null = null;

  try {
    const fetched = await fetchImageBlob(srcUrl);

    if (typeof OffscreenCanvas !== "undefined") {
      thumbnailBlob = await generateThumbnail(fetched.blob);
    }

    imageId = await saveImage({
      blob: fetched.blob,
      mimeType: fetched.mimeType,
      title,
      pageUrl,
      accessToken: auth.accessToken,
      thumbnailBlob: thumbnailBlob ?? undefined,
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

import { getAuth, type BookleafAuth } from "../lib/storage";
import { apiFetch } from "../lib/api";

chrome.runtime.onInstalled.addListener(() => {
  chrome.contextMenus.removeAll(() => {
    chrome.contextMenus.create({
      id: "save-to-bookleaf",
      title: "Save to Bookleaf",
      contexts: ["image"],
    });
  });
});

chrome.contextMenus.onClicked.addListener((info, tab) => {
  if (info.menuItemId !== "save-to-bookleaf") return;
  const srcUrl = info.srcUrl;
  const pageUrl = info.pageUrl;
  const title = tab?.title ?? pageUrl ?? "Untitled";
  if (!srcUrl) return;
  handleSave({ srcUrl, pageUrl: pageUrl ?? "", title });
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

async function saveImage({
  blob,
  mimeType,
  title,
  pageUrl,
  accessToken,
}: {
  blob: Blob;
  mimeType: string;
  title: string;
  pageUrl: string;
  accessToken: string;
}): Promise<void> {
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

  const { upload_url, id: image_id } = (await initRes.json()) as {
    upload_url: string;
    id: string;
  };

  const putRes = await fetch(upload_url, {
    method: "PUT",
    headers: { "Content-Type": mimeType },
    body: blob,
  });
  if (!putRes.ok) throw new Error(`PUT to R2 failed: ${putRes.status}`);

  const completeRes = await apiFetch(`/images/${image_id}/complete`, {
    method: "POST",
  });
  if (!completeRes.ok)
    throw new Error(`POST /complete failed: ${completeRes.status}`);
}

async function handleSave({
  srcUrl,
  pageUrl,
  title,
}: {
  srcUrl: string;
  pageUrl: string;
  title: string;
}): Promise<void> {
  const auth = await getAuth();

  if (!isTokenValid(auth)) {
    notify("Bookleaf", "Please log in first");
    return;
  }

  try {
    const { blob, mimeType } = await fetchImageBlob(srcUrl);
    await saveImage({
      blob,
      mimeType,
      title,
      pageUrl,
      accessToken: auth.accessToken,
    });
    notify("Bookleaf", "Saved to Bookleaf!");
  } catch {
    notify("Bookleaf", "Save failed. Please try again.");
  }
}

function notify(title: string, message: string): void {
  chrome.notifications.create({
    type: "basic",
    iconUrl: chrome.runtime.getURL("icons/icon48.png"),
    title,
    message,
  });
}

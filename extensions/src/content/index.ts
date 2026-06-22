import browser from "webextension-polyfill";
import { resolveCardImageSrc, shouldResolveCardDom } from "../lib/cardDomResolveRules";
import { isTwitterUrl, resolveTweetText } from "../lib/tweetTextResolveRule";
import { isImgurUrl, isInstagramUrl, resolveAltText } from "../lib/altTextResolveRule";
import { isFacebookUrl, resolveFacebookAltText } from "../lib/facebookAltResolveRule";
import { resolveDragSaveContext, type DragSaveContext } from "../lib/dragSaveContext";
import { getDragEnabled } from "../lib/storage";

type ToastVariant = "success" | "error";

interface ToastMessage {
  type: "toast";
  variant: ToastVariant;
  title: string;
  body: string;
}

const host = document.createElement("div");
host.id = "bookleaf-toast-host";
document.body.appendChild(host);

const shadow = host.attachShadow({ mode: "open" });

const style = document.createElement("style");
style.textContent = `
  .toast {
    position: fixed;
    bottom: 24px;
    right: 24px;
    z-index: 2147483647;
    display: flex;
    flex-direction: column;
    gap: 2px;
    min-width: 260px;
    max-width: 360px;
    padding: 14px 16px;
    background: #ffffff;
    border-radius: 8px;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15), 0 1px 3px rgba(0, 0, 0, 0.1);
    border-left: 4px solid transparent;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
    font-size: 14px;
    line-height: 1.4;
    animation: slide-in 0.2s ease-out;
  }

  .toast.success {
    border-left-color: #22c55e;
  }

  .toast.error {
    border-left-color: #ef4444;
  }

  .toast-title {
    font-weight: 600;
    color: #0a0a0a;
  }

  .toast-body {
    font-weight: 400;
    color: #555555;
  }

  .toast.fade-out {
    animation: fade-out 0.3s ease-in forwards;
  }

  @keyframes slide-in {
    from {
      opacity: 0;
      transform: translateY(8px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  @keyframes fade-out {
    from {
      opacity: 1;
    }
    to {
      opacity: 0;
    }
  }

  .drop-zone {
    position: fixed;
    z-index: 2147483647;
    width: 150px;
    height: 150px;
    margin: -75px 0 0 -75px;
    box-sizing: border-box;
    padding: 8px;
    border-radius: 8px;
    background: #faf8f4;
    box-shadow: 0 8px 20px rgba(0, 0, 0, 0.12), 0 2px 6px rgba(0, 0, 0, 0.08);
    pointer-events: auto;
  }

  .drop-zone-border {
    width: 100%;
    height: 100%;
    box-sizing: border-box;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 8px;
    padding: 8px;
    border: 2px dashed rgba(140, 132, 123, 0.3);
    border-radius: 6px;
    color: #8c847b;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
    font-size: 12px;
    font-weight: 500;
    text-align: center;
  }

  .drop-zone-icon {
    width: 28px;
    height: 28px;
  }

  .snip-overlay {
    position: fixed;
    inset: 0;
    z-index: 2147483647;
    cursor: crosshair;
    overflow: hidden;
    user-select: none;
  }

  .snip-overlay img {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    display: block;
    pointer-events: none;
  }

  .snip-rect {
    position: absolute;
    box-shadow: 0 0 0 9999px rgba(0, 0, 0, 0.5);
    outline: 1px solid #ffffff;
    pointer-events: none;
  }
`;

shadow.appendChild(style);

let currentToast: HTMLElement | null = null;
let dismissTimer: ReturnType<typeof setTimeout> | null = null;

function showToast(variant: ToastVariant, title: string, body: string): void {
  if (currentToast) {
    currentToast.remove();
    currentToast = null;
  }
  if (dismissTimer !== null) {
    clearTimeout(dismissTimer);
    dismissTimer = null;
  }

  const toast = document.createElement("div");
  toast.className = `toast ${variant}`;

  const titleEl = document.createElement("span");
  titleEl.className = "toast-title";
  titleEl.textContent = title;

  const bodyEl = document.createElement("span");
  bodyEl.className = "toast-body";
  bodyEl.textContent = body;

  toast.appendChild(titleEl);
  toast.appendChild(bodyEl);
  shadow.appendChild(toast);
  currentToast = toast;

  dismissTimer = setTimeout(() => {
    if (!currentToast) return;
    currentToast.classList.add("fade-out");
    currentToast.addEventListener("animationend", () => {
      currentToast?.remove();
      currentToast = null;
    }, { once: true });
    dismissTimer = null;
  }, 4000);
}

interface SnipFrameMessage {
  type: "snip-frame";
  dataUrl: string;
}

interface CaptureVideoFrameMessage {
  type: "capture-video-frame";
}

let lastRightClickedVideo: HTMLVideoElement | null = null;

async function captureVideoFrame(): Promise<void> {
  const video = lastRightClickedVideo;
  if (!video || video.videoWidth === 0) {
    showToast("error", "Bookleaf", "Can't capture this video.");
    return;
  }

  try {
    const canvas = document.createElement("canvas");
    canvas.width = video.videoWidth;
    canvas.height = video.videoHeight;
    const ctx = canvas.getContext("2d");
    if (!ctx) throw new Error("Could not get 2d context");
    ctx.drawImage(video, 0, 0, video.videoWidth, video.videoHeight);

    const blob = await new Promise<Blob | null>((resolve) => canvas.toBlob(resolve, "image/png"));
    if (!blob) throw new Error("toBlob returned null");

    browser.runtime.sendMessage({ type: "video-frame-captured", blob, mimeType: "image/png" });
  } catch {
    showToast("error", "Bookleaf", "Can't capture this video.");
  }
}

browser.runtime.onMessage.addListener((message: unknown) => {
  const msg = message as ToastMessage | SnipFrameMessage | CaptureVideoFrameMessage;
  if (msg.type === "toast") {
    showToast(msg.variant, msg.title, msg.body);
    return;
  }
  if (msg.type === "snip-frame") {
    renderSnipOverlay(msg.dataUrl);
    return;
  }
  if (msg.type === "capture-video-frame") {
    void captureVideoFrame();
  }
});

let snipOverlay: HTMLElement | null = null;
let snipImage: HTMLImageElement | null = null;
let snipRect: HTMLElement | null = null;
let snipStart: { x: number; y: number } | null = null;
let snipDragging = false;

function removeSnipOverlay(): void {
  snipOverlay?.remove();
  snipOverlay = null;
  snipImage = null;
  snipRect = null;
  snipStart = null;
  snipDragging = false;
  document.removeEventListener("keydown", handleSnipKeydown, true);
}

function handleSnipKeydown(event: KeyboardEvent): void {
  if (event.key !== "Escape") return;
  removeSnipOverlay();
}

function updateSnipRect(clientX: number, clientY: number): void {
  if (!snipStart || !snipRect) return;
  const left = Math.min(snipStart.x, clientX);
  const top = Math.min(snipStart.y, clientY);
  const width = Math.abs(clientX - snipStart.x);
  const height = Math.abs(clientY - snipStart.y);
  snipRect.style.left = `${left}px`;
  snipRect.style.top = `${top}px`;
  snipRect.style.width = `${width}px`;
  snipRect.style.height = `${height}px`;
}

function handleSnipMouseDown(event: MouseEvent): void {
  snipStart = { x: event.clientX, y: event.clientY };
  snipDragging = true;
  updateSnipRect(event.clientX, event.clientY);
}

function handleSnipMouseMove(event: MouseEvent): void {
  if (!snipDragging) return;
  updateSnipRect(event.clientX, event.clientY);
}

function handleSnipMouseUp(): void {
  if (!snipDragging) return;
  snipDragging = false;
  void finalizeSnipSelection();
}

async function finalizeSnipSelection(): Promise<void> {
  const image = snipImage;
  const rect = snipRect;

  if (image && rect) {
    const bounds = rect.getBoundingClientRect();
    if (bounds.width > 0 && bounds.height > 0) {
      const scaleX = image.naturalWidth / window.innerWidth;
      const scaleY = image.naturalHeight / window.innerHeight;

      const canvas = document.createElement("canvas");
      canvas.width = Math.round(bounds.width * scaleX);
      canvas.height = Math.round(bounds.height * scaleY);
      const ctx = canvas.getContext("2d");

      if (ctx) {
        ctx.drawImage(
          image,
          bounds.left * scaleX,
          bounds.top * scaleY,
          bounds.width * scaleX,
          bounds.height * scaleY,
          0,
          0,
          canvas.width,
          canvas.height,
        );
        const blob = await new Promise<Blob | null>((resolve) =>
          canvas.toBlob(resolve, "image/png"),
        );
        if (blob) {
          browser.runtime.sendMessage({ type: "snip-captured", blob, mimeType: "image/png" });
        }
      }
    }
  }

  removeSnipOverlay();
}

function renderSnipOverlay(dataUrl: string): void {
  removeSnipOverlay();

  const overlay = document.createElement("div");
  overlay.className = "snip-overlay";

  const image = document.createElement("img");
  image.src = dataUrl;

  const rect = document.createElement("div");
  rect.className = "snip-rect";
  rect.style.left = "0px";
  rect.style.top = "0px";
  rect.style.width = "0px";
  rect.style.height = "0px";

  overlay.appendChild(image);
  overlay.appendChild(rect);
  shadow.appendChild(overlay);

  snipOverlay = overlay;
  snipImage = image;
  snipRect = rect;
  snipStart = null;
  snipDragging = false;

  overlay.addEventListener("mousedown", handleSnipMouseDown);
  overlay.addEventListener("mousemove", handleSnipMouseMove);
  overlay.addEventListener("mouseup", handleSnipMouseUp);
  document.addEventListener("keydown", handleSnipKeydown, true);
}

document.addEventListener(
  "contextmenu",
  (event) => {
    if (!(event.target instanceof Element)) return;

    lastRightClickedVideo =
      event.target instanceof HTMLVideoElement ? event.target : event.target.querySelector("video");

    const resolved: Partial<{ srcUrl: string; title: string }> = {};

    if (shouldResolveCardDom(window.location.href)) {
      const srcUrl = resolveCardImageSrc(event.target);
      if (srcUrl) resolved.srcUrl = srcUrl;
    }

    if (isTwitterUrl(window.location.href)) {
      const title = resolveTweetText(event.target);
      if (title) resolved.title = title;
    }

    if (isImgurUrl(window.location.href) || isInstagramUrl(window.location.href)) {
      const title = resolveAltText(event.target);
      if (title) resolved.title = title;
    }

    if (isFacebookUrl(window.location.href)) {
      const title = resolveFacebookAltText(event.target);
      if (title) resolved.title = title;
    }

    if (
      !shouldResolveCardDom(window.location.href) &&
      !isTwitterUrl(window.location.href) &&
      !isImgurUrl(window.location.href) &&
      !isInstagramUrl(window.location.href) &&
      !isFacebookUrl(window.location.href)
    )
      return;
    browser.runtime.sendMessage({ resolved });
  },
  { capture: true },
);

function isBookleafAppUrl(pageUrl: string): boolean {
  const appUrl = import.meta.env.VITE_APP_URL;
  if (!appUrl) return false;
  try {
    return new URL(pageUrl).origin === new URL(appUrl).origin;
  } catch {
    return false;
  }
}

let dragContext: DragSaveContext | null = null;
let dropZone: HTMLElement | null = null;

function removeDropZone(): void {
  dropZone?.remove();
  dropZone = null;
}

const UPLOAD_CLOUD_ICON_SVG = `
  <svg class="drop-zone-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
    <path d="M12 13v8" />
    <path d="M4 14.899A7 7 0 1 1 15.71 8h1.79a4.5 4.5 0 0 1 2.5 8.242" />
    <path d="m8 17 4-4 4 4" />
  </svg>
`;

function renderDropZone(clientX: number, clientY: number): void {
  const zone = document.createElement("div");
  zone.className = "drop-zone";
  zone.style.left = `${clientX}px`;
  zone.style.top = `${clientY}px`;

  const border = document.createElement("div");
  border.className = "drop-zone-border";
  border.innerHTML = UPLOAD_CLOUD_ICON_SVG;

  const label = document.createElement("span");
  label.textContent = "Save to Bookleaf";
  border.appendChild(label);

  zone.appendChild(border);

  zone.addEventListener("dragover", (event) => {
    event.preventDefault();
  });

  zone.addEventListener("drop", (event) => {
    event.preventDefault();
    if (dragContext) {
      browser.runtime.sendMessage({ type: "drag-save", ...dragContext });
    }
  });

  shadow.appendChild(zone);
  dropZone = zone;
}

document.addEventListener(
  "dragstart",
  (event) => {
    if (!(event.target instanceof Element)) return;

    const pageUrl = window.location.href;
    if (isBookleafAppUrl(pageUrl)) return;

    const target = event.target;
    const clientX = event.clientX;
    const clientY = event.clientY;

    void getDragEnabled().then((dragEnabled) => {
      const context = resolveDragSaveContext(target, pageUrl, dragEnabled);
      dragContext = context;
      if (context) renderDropZone(clientX, clientY);
    });
  },
  { capture: true },
);

function endDrag(): void {
  dragContext = null;
  removeDropZone();
}

document.addEventListener("dragend", endDrag, { capture: true });

// Safety net: some sites (e.g. dnd-kit/react-dnd-based drag UIs) call preventDefault()
// on dragstart to cancel native drag in favor of synthetic pointer-based dragging, which
// means dragend never fires natively. pointerup/mouseup always fire once the gesture ends,
// regardless of whether native drag was cancelled, so they catch the case dragend misses.
document.addEventListener("pointerup", endDrag, { capture: true });
document.addEventListener("mouseup", endDrag, { capture: true });

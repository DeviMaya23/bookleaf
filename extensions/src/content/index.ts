import browser from "webextension-polyfill";
import { resolveCardImageSrc, shouldResolveCardDom } from "../lib/cardDomResolveRules";
import { isTwitterUrl, resolveTweetText } from "../lib/tweetTextResolveRule";
import { isImgurUrl, isInstagramUrl, resolveAltText } from "../lib/altTextResolveRule";
import { isFacebookUrl, resolveFacebookAltText } from "../lib/facebookAltResolveRule";
import { resolveDragSaveContext, type DragSaveContext } from "../lib/dragSaveContext";
import { getDarkMode, getDragEnabled } from "../lib/storage";
import { collectPageImages, type PageImage } from "../lib/imagePicker";

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
  :host([data-theme="light"]) {
    --bl-bg: #ffffff;
    --bl-text: #0a0a0a;
    --bl-text-sec: #555555;
    --bl-border: rgba(140, 132, 123, 0.3);
    --bl-accent: #22c55e;
    --bl-surface: #faf8f4;
    --bl-success: #22c55e;
    --bl-error: #ef4444;
  }

  :host([data-theme="dark"]) {
    --bl-bg: #1a1a1a;
    --bl-text: #f0f0f0;
    --bl-text-sec: #a0a0a0;
    --bl-border: rgba(255, 255, 255, 0.15);
    --bl-accent: #22c55e;
    --bl-surface: #2a2a2a;
    --bl-success: #22c55e;
    --bl-error: #ef4444;
  }

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
    background: var(--bl-bg);
    border-radius: 8px;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15), 0 1px 3px rgba(0, 0, 0, 0.1);
    border-left: 4px solid transparent;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
    font-size: 14px;
    line-height: 1.4;
    animation: slide-in 0.2s ease-out;
  }

  .toast.success {
    border-left-color: var(--bl-success);
  }

  .toast.error {
    border-left-color: var(--bl-error);
  }

  .toast-title {
    font-weight: 600;
    color: var(--bl-text);
  }

  .toast-body {
    font-weight: 400;
    color: var(--bl-text-sec);
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
    background: var(--bl-surface);
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
    border: 2px dashed var(--bl-border);
    border-radius: 6px;
    color: var(--bl-text-sec);
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

  .picker-backdrop {
    position: fixed;
    inset: 0;
    z-index: 2147483646;
    background: rgba(0, 0, 0, 0.6);
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .picker-panel {
    background: var(--bl-bg);
    border-radius: 12px;
    box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
    width: 680px;
    max-width: 90vw;
    max-height: 80vh;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
  }

  .picker-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 16px 20px;
    border-bottom: 1px solid var(--bl-border);
  }

  .picker-header-label {
    font-size: 14px;
    font-weight: 600;
    color: var(--bl-text);
  }

  .picker-close {
    background: transparent;
    border: none;
    cursor: pointer;
    color: var(--bl-text-sec);
    font-size: 18px;
    line-height: 1;
    padding: 2px 6px;
  }

  .picker-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
    grid-auto-rows: minmax(120px, 120px);
    gap: 10px;
    padding: 16px;
    overflow-y: auto;
    flex: 1;
  }

  .picker-thumb {
    position: relative;
    border-radius: 6px;
    overflow: hidden;
    cursor: pointer;
    border: 2px solid transparent;
    background: var(--bl-surface);
  }

  .picker-thumb.selected {
    border-color: var(--bl-accent);
  }

  .picker-thumb img {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
    pointer-events: none;
  }

  .picker-thumb-checkbox {
    position: absolute;
    top: 6px;
    right: 6px;
    width: 18px;
    height: 18px;
    border-radius: 50%;
    border: 2px solid rgba(255, 255, 255, 0.8);
    background: rgba(0, 0, 0, 0.3);
    display: flex;
    align-items: center;
    justify-content: center;
    pointer-events: none;
  }

  .picker-thumb.selected .picker-thumb-checkbox {
    background: var(--bl-accent);
    border-color: var(--bl-accent);
  }

  .picker-thumb.selected .picker-thumb-checkbox::after {
    content: '';
    width: 5px;
    height: 9px;
    border: 2px solid #fff;
    border-top: none;
    border-left: none;
    transform: rotate(45deg) translate(-1px, -1px);
    display: block;
  }

  .picker-thumb-dim {
    position: absolute;
    bottom: 0;
    left: 0;
    right: 0;
    padding: 2px 4px;
    font-size: 10px;
    color: rgba(255, 255, 255, 0.9);
    background: rgba(0, 0, 0, 0.45);
    text-align: center;
    pointer-events: none;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
  }

  .picker-footer {
    padding: 14px 20px;
    border-top: 1px solid var(--bl-border);
    display: flex;
    justify-content: flex-end;
  }

  .picker-confirm {
    background: var(--bl-accent);
    color: #ffffff;
    border: none;
    border-radius: 8px;
    padding: 9px 20px;
    font-size: 14px;
    font-weight: 600;
    cursor: pointer;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
  }

  .picker-confirm:disabled {
    opacity: 0.4;
    cursor: default;
  }
`;

shadow.appendChild(style);

async function applyTheme(hostEl: HTMLElement): Promise<void> {
  const dark = await getDarkMode();
  hostEl.dataset.theme = dark ? "dark" : "light";
}

let currentToast: HTMLElement | null = null;
let dismissTimer: ReturnType<typeof setTimeout> | null = null;

async function showToast(variant: ToastVariant, title: string, body: string): Promise<void> {
  await applyTheme(host);

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
    void showToast("error", "Bookleaf", "Can't capture this video.");
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
    void showToast("error", "Bookleaf", "Can't capture this video.");
  }
}

interface OpenImagePickerMessage {
  type: "open-image-picker";
}

browser.runtime.onMessage.addListener((message: unknown) => {
  const msg = message as ToastMessage | SnipFrameMessage | CaptureVideoFrameMessage | OpenImagePickerMessage;
  if (msg.type === "toast") {
    void showToast(msg.variant, msg.title, msg.body);
    return;
  }
  if (msg.type === "snip-frame") {
    renderSnipOverlay(msg.dataUrl);
    return;
  }
  if (msg.type === "capture-video-frame") {
    void captureVideoFrame();
    return;
  }
  if (msg.type === "open-image-picker") {
    void openImagePicker();
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

async function renderDropZone(clientX: number, clientY: number): Promise<void> {
  await applyTheme(host);

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
      if (context) void renderDropZone(clientX, clientY);
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

let pickerOverlay: HTMLElement | null = null;
let pickerEscHandler: ((e: KeyboardEvent) => void) | null = null;

function removePickerOverlay(): void {
  if (!pickerOverlay) return;
  pickerOverlay.remove();
  pickerOverlay = null;
  if (pickerEscHandler) {
    document.removeEventListener("keydown", pickerEscHandler, true);
    pickerEscHandler = null;
  }
}

function formatDimensions(image: PageImage): string {
  if (image.srcsetWidth !== null && image.naturalWidth > 0) {
    const h = Math.round(image.srcsetWidth * image.naturalHeight / image.naturalWidth);
    return `${image.srcsetWidth} × ${h}`;
  }
  return `${image.naturalWidth} × ${image.naturalHeight}`;
}

function renderPickerOverlay(images: PageImage[]): void {
  removePickerOverlay();

  const selected = new Set<string>();

  const backdrop = document.createElement("div");
  backdrop.className = "picker-backdrop";

  const panel = document.createElement("div");
  panel.className = "picker-panel";

  // Header
  const header = document.createElement("div");
  header.className = "picker-header";

  const headerLabel = document.createElement("span");
  headerLabel.className = "picker-header-label";
  headerLabel.textContent = `${images.length} ${images.length === 1 ? "image" : "images"} found`;

  const closeBtn = document.createElement("button");
  closeBtn.className = "picker-close";
  closeBtn.textContent = "✕";
  closeBtn.addEventListener("click", removePickerOverlay);

  header.appendChild(headerLabel);
  header.appendChild(closeBtn);

  // Grid
  const grid = document.createElement("div");
  grid.className = "picker-grid";

  // Footer
  const footer = document.createElement("div");
  footer.className = "picker-footer";

  const confirmBtn = document.createElement("button");
  confirmBtn.className = "picker-confirm";
  confirmBtn.textContent = "Save 0 images";
  confirmBtn.disabled = true;

  footer.appendChild(confirmBtn);

  function updateConfirmBtn(): void {
    const count = selected.size;
    confirmBtn.disabled = count === 0;
    confirmBtn.textContent = count === 1 ? "Save 1 image" : `Save ${count} images`;
  }

  for (const image of images) {
    const card = document.createElement("div");
    card.className = "picker-thumb";

    const img = document.createElement("img");
    img.src = image.src;
    img.alt = "";

    const dim = document.createElement("span");
    dim.className = "picker-thumb-dim";
    dim.textContent = formatDimensions(image);

    const checkbox = document.createElement("div");
    checkbox.className = "picker-thumb-checkbox";

    card.appendChild(img);
    card.appendChild(dim);
    card.appendChild(checkbox);

    card.addEventListener("click", () => {
      if (selected.has(image.src)) {
        selected.delete(image.src);
        card.classList.remove("selected");
      } else {
        selected.add(image.src);
        card.classList.add("selected");
      }
      updateConfirmBtn();
    });

    grid.appendChild(card);
  }

  confirmBtn.addEventListener("click", () => {
    const imgs = [...selected].map((srcUrl) => ({ srcUrl }));
    removePickerOverlay();
    browser.runtime.sendMessage({ type: "picker-save", images: imgs });
  });

  panel.appendChild(header);
  panel.appendChild(grid);
  panel.appendChild(footer);
  backdrop.appendChild(panel);
  backdrop.addEventListener("click", (e) => {
    if (e.target === backdrop) removePickerOverlay();
  });
  shadow.appendChild(backdrop);
  pickerOverlay = backdrop;

  pickerEscHandler = (e: KeyboardEvent) => {
    if (e.key === "Escape") removePickerOverlay();
  };
  document.addEventListener("keydown", pickerEscHandler, true);
}

async function openImagePicker(): Promise<void> {
  if (isBookleafAppUrl(window.location.href)) return;
  const images = collectPageImages(document);
  if (images.length === 0) {
    await showToast("error", "Bookleaf", "No images found on this page.");
    return;
  }
  await applyTheme(host);
  renderPickerOverlay(images);
}

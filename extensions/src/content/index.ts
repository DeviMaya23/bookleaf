import browser from "webextension-polyfill";
import { resolveCardImageSrc, shouldResolveCardDom } from "../lib/cardDomResolveRules";
import { isTwitterUrl, resolveTweetText } from "../lib/tweetTextResolveRule";
import { isImgurUrl, isInstagramUrl, resolveAltText } from "../lib/altTextResolveRule";
import { isFacebookUrl, resolveFacebookAltText } from "../lib/facebookAltResolveRule";
import { resolveDragImageSrc } from "../lib/dragImageResolveRule";

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
    width: 96px;
    height: 96px;
    margin: -48px 0 0 -48px;
    display: flex;
    align-items: center;
    justify-content: center;
    border: 2px dashed #6366f1;
    border-radius: 12px;
    background: rgba(99, 102, 241, 0.1);
    color: #6366f1;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
    font-size: 12px;
    font-weight: 600;
    text-align: center;
    pointer-events: auto;
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

browser.runtime.onMessage.addListener((message: unknown) => {
  const msg = message as ToastMessage;
  if (msg.type !== "toast") return;
  showToast(msg.variant, msg.title, msg.body);
});

document.addEventListener(
  "contextmenu",
  (event) => {
    if (!(event.target instanceof Element)) return;

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

interface DragSaveContext {
  srcUrl: string;
  title?: string;
  linkUrl?: string;
}

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

function resolveDragTitle(target: Element): string | undefined {
  const pageUrl = window.location.href;
  if (isTwitterUrl(pageUrl)) return resolveTweetText(target) ?? undefined;
  if (isImgurUrl(pageUrl) || isInstagramUrl(pageUrl)) return resolveAltText(target) ?? undefined;
  if (isFacebookUrl(pageUrl)) return resolveFacebookAltText(target) ?? undefined;
  return undefined;
}

function removeDropZone(): void {
  dropZone?.remove();
  dropZone = null;
}

function renderDropZone(clientX: number, clientY: number): void {
  const zone = document.createElement("div");
  zone.className = "drop-zone";
  zone.textContent = "Drop to save";
  zone.style.left = `${clientX}px`;
  zone.style.top = `${clientY}px`;

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

    const srcUrl = resolveDragImageSrc(event.target, pageUrl);
    if (!srcUrl) {
      dragContext = null;
      return;
    }

    dragContext = {
      srcUrl,
      title: resolveDragTitle(event.target),
      linkUrl: event.target.closest("a")?.href,
    };

    renderDropZone(event.clientX, event.clientY);
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

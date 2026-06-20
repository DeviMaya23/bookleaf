import browser from "webextension-polyfill";
import { resolveCardImageSrc, shouldResolveCardDom } from "../lib/cardDomResolveRules";
import { isTwitterUrl, resolveTweetText } from "../lib/tweetTextResolveRule";

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

    if (!shouldResolveCardDom(window.location.href) && !isTwitterUrl(window.location.href)) return;
    browser.runtime.sendMessage({ resolved });
  },
  { capture: true },
);

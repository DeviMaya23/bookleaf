import { vi } from "vitest";

export function createBrowserMock() {
  const store = new Map<string, unknown>();

  return {
    action: {
      setBadgeText: vi.fn(),
      setBadgeBackgroundColor: vi.fn(),
    },
    identity: {
      getRedirectURL: vi.fn(() => "https://redirect.test"),
      launchWebAuthFlow: vi.fn(async () => null as string | null),
    },
    storage: {
      local: {
        get: vi.fn(async (key: string) => ({ [key]: store.get(key) })),
        set: vi.fn(async (items: Record<string, unknown>) => {
          for (const [key, value] of Object.entries(items)) store.set(key, value);
        }),
        remove: vi.fn(async (key: string) => {
          store.delete(key);
        }),
      },
    },
    tabs: {
      sendMessage: vi.fn(async () => undefined),
    },
    contextMenus: {
      removeAll: vi.fn(async () => undefined),
      create: vi.fn(),
      onClicked: {
        addListener: vi.fn(),
      },
    },
    runtime: {
      onInstalled: {
        addListener: vi.fn(),
      },
    },
  };
}

import browser from "webextension-polyfill";

export interface BookleafAuth {
  accessToken: string;
  refreshToken?: string;
  expiresAt: number;
}

export interface RecentSave {
  imageId: string;
  title: string;
  dataUrl: string;
  savedAt: number;
}

const STORAGE_KEY = "bookleaf_auth";
const RECENT_SAVES_KEY = "bookleaf_recent_saves";
const DARK_MODE_KEY = "bookleaf_dark_mode";
const DRAG_ENABLED_KEY = "bookleaf_drag_enabled";
const USERNAME_KEY = "bookleaf_username";
const AVATAR_KEY = "bookleaf_avatar";

const MAX_RECENT_SAVES = 5;

export async function getAuth(): Promise<BookleafAuth | null> {
  const result = await browser.storage.local.get(STORAGE_KEY);
  return (result[STORAGE_KEY] as BookleafAuth) ?? null;
}

export async function setAuth(auth: BookleafAuth): Promise<void> {
  await browser.storage.local.set({ [STORAGE_KEY]: auth });
}

export async function clearAuth(): Promise<void> {
  await browser.storage.local.remove(STORAGE_KEY);
}

export async function getRecentSaves(): Promise<RecentSave[]> {
  const result = await browser.storage.local.get(RECENT_SAVES_KEY);
  return (result[RECENT_SAVES_KEY] as RecentSave[]) ?? [];
}

export async function addRecentSave(entry: RecentSave): Promise<void> {
  const current = await getRecentSaves();
  const updated = [entry, ...current].slice(0, MAX_RECENT_SAVES);
  await browser.storage.local.set({ [RECENT_SAVES_KEY]: updated });
}

export async function getDarkMode(): Promise<boolean> {
  const result = await browser.storage.local.get(DARK_MODE_KEY);
  return (result[DARK_MODE_KEY] as boolean) ?? false;
}

export async function setDarkMode(value: boolean): Promise<void> {
  await browser.storage.local.set({ [DARK_MODE_KEY]: value });
}

export async function getDragEnabled(): Promise<boolean> {
  const result = await browser.storage.local.get(DRAG_ENABLED_KEY);
  return (result[DRAG_ENABLED_KEY] as boolean) ?? true;
}

export async function setDragEnabled(value: boolean): Promise<void> {
  await browser.storage.local.set({ [DRAG_ENABLED_KEY]: value });
}

export async function getUsername(): Promise<string | null> {
  const result = await browser.storage.local.get(USERNAME_KEY);
  return (result[USERNAME_KEY] as string) ?? null;
}

export async function setUsername(name: string): Promise<void> {
  await browser.storage.local.set({ [USERNAME_KEY]: name });
}

export async function getAvatar(): Promise<string | null> {
  const result = await browser.storage.local.get(AVATAR_KEY);
  return (result[AVATAR_KEY] as string) ?? null;
}

export async function setAvatar(url: string): Promise<void> {
  await browser.storage.local.set({ [AVATAR_KEY]: url });
}

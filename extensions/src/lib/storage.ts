import browser from "webextension-polyfill";

export interface BookleafAuth {
  accessToken: string;
  refreshToken?: string;
  expiresAt: number;
}

const STORAGE_KEY = "bookleaf_auth";

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

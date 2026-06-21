import { describe, expect, it, vi } from "vitest";

vi.mock("webextension-polyfill", async () => {
  const { createBrowserMock } = await import("../test/browserMock");
  return { default: createBrowserMock() };
});

import { addRecentSave, getDragEnabled, getRecentSaves, setDragEnabled, type RecentSave } from "./storage";

function makeSave(imageId: string): RecentSave {
  return { imageId, title: imageId, dataUrl: "data:image/jpeg;base64,", savedAt: Date.now() };
}

describe("addRecentSave", () => {
  it("prepends the newest entry and caps the list at 5", async () => {
    for (let i = 0; i < 5; i++) {
      await addRecentSave(makeSave(`old-${i}`));
    }

    await addRecentSave(makeSave("newest"));

    const saves = await getRecentSaves();
    expect(saves).toHaveLength(5);
    expect(saves[0].imageId).toBe("newest");
  });
});

describe("getDragEnabled/setDragEnabled", () => {
  it("returns true when no value has ever been stored", async () => {
    expect(await getDragEnabled()).toBe(true);
  });

  it("returns the last-set value after setDragEnabled(false)", async () => {
    await setDragEnabled(false);
    expect(await getDragEnabled()).toBe(false);
  });

  it("returns the last-set value after setDragEnabled(true)", async () => {
    await setDragEnabled(false);
    await setDragEnabled(true);
    expect(await getDragEnabled()).toBe(true);
  });
});

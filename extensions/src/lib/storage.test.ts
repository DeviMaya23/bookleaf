import { describe, expect, it, vi } from "vitest";

vi.mock("webextension-polyfill", async () => {
  const { createBrowserMock } = await import("../test/browserMock");
  return { default: createBrowserMock() };
});

import { addRecentSave, getRecentSaves, type RecentSave } from "./storage";

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

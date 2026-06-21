import { describe, expect, it, vi } from "vitest";

vi.mock("webextension-polyfill", async () => {
  const { createBrowserMock } = await import("../test/browserMock");
  return { default: createBrowserMock() };
});

import browser from "webextension-polyfill";
import { openChromeShortcutSettings, openFirefoxShortcutSettings } from "./Settings";

describe("openFirefoxShortcutSettings (Firefox path)", () => {
  it("opens the browser's native shortcut settings and does not attempt browser.commands.update", async () => {
    await openFirefoxShortcutSettings();

    expect(browser.commands.openShortcutSettings).toHaveBeenCalled();
    expect(browser.commands.update).not.toHaveBeenCalled();
  });
});

describe("openChromeShortcutSettings (Chrome path)", () => {
  it("opens chrome://extensions/shortcuts and does not attempt browser.commands.update", async () => {
    await openChromeShortcutSettings();

    expect(browser.tabs.create).toHaveBeenCalledWith({ url: "chrome://extensions/shortcuts" });
    expect(browser.commands.update).not.toHaveBeenCalled();
  });
});

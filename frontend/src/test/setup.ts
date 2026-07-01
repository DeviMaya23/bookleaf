import '@testing-library/jest-dom'

if (!window.matchMedia) {
  window.matchMedia = (query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addEventListener: () => {},
    removeEventListener: () => {},
    addListener: () => {},
    removeListener: () => {},
    dispatchEvent: () => false,
  }) as unknown as MediaQueryList
}

globalThis.ResizeObserver = class ResizeObserver {
  private callback: ResizeObserverCallback
  constructor(cb: ResizeObserverCallback) {
    this.callback = cb
  }
  observe(target: Element) {
    this.callback(
      [{ contentRect: { width: 800 } } as ResizeObserverEntry],
      this,
    )
    Object.defineProperty(target, 'getBoundingClientRect', {
      configurable: true,
      value: () => ({ width: 800 }),
    })
  }
  unobserve() {}
  disconnect() {}
}

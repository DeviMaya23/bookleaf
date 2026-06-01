import '@testing-library/jest-dom'

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

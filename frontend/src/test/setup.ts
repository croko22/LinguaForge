import '@testing-library/jest-dom/vitest'

class ResizeObserverMock {
  observe() {}
  unobserve() {}
  disconnect() {}
}
// @ts-expect-error ResizeObserver not available in jsdom
globalThis.ResizeObserver = ResizeObserverMock

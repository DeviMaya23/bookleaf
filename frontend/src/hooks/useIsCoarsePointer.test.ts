import { describe, it, expect, afterEach } from 'vitest'
import { act, renderHook } from '@testing-library/react'
import { useIsCoarsePointer } from './useIsCoarsePointer'

function mockMatchMedia(initialMatches: boolean) {
  const listeners = new Set<(event: MediaQueryListEvent) => void>()
  let matches = initialMatches
  const mql: Partial<MediaQueryList> = {
    get matches() {
      return matches
    },
    media: '(pointer: coarse)',
    addEventListener: (_type: string, listener: EventListenerOrEventListenerObject) => {
      listeners.add(listener as (event: MediaQueryListEvent) => void)
    },
    removeEventListener: (_type: string, listener: EventListenerOrEventListenerObject) => {
      listeners.delete(listener as (event: MediaQueryListEvent) => void)
    },
  }
  window.matchMedia = () => mql as MediaQueryList
  return {
    fireChange: (newMatches: boolean) => {
      matches = newMatches
      for (const listener of listeners) listener({ matches: newMatches } as MediaQueryListEvent)
    },
  }
}

describe('useIsCoarsePointer', () => {
  const originalMatchMedia = window.matchMedia

  afterEach(() => {
    window.matchMedia = originalMatchMedia
  })

  it('reflects the initial matchMedia result', () => {
    mockMatchMedia(true)

    const { result } = renderHook(() => useIsCoarsePointer())

    expect(result.current).toBe(true)
  })

  it('updates when the media query change event fires', () => {
    const { fireChange } = mockMatchMedia(false)

    const { result } = renderHook(() => useIsCoarsePointer())
    expect(result.current).toBe(false)

    act(() => fireChange(true))

    expect(result.current).toBe(true)
  })
})

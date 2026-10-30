import { renderHook, act } from '@testing-library/react'
import { vi, describe, it, expect, beforeEach, afterEach } from 'vitest'
import { useDebouncedValue } from './useDebouncedValue'

describe('useDebouncedValue', () => {
  beforeEach(() => { vi.useFakeTimers() })
  afterEach(() => { vi.useRealTimers() })

  it('returns the updated value after the delay elapses', () => {
    const { result, rerender } = renderHook(
      ({ value }) => useDebouncedValue(value, 300),
      { initialProps: { value: 'a' } },
    )
    expect(result.current).toBe('a')
    rerender({ value: 'b' })
    expect(result.current).toBe('a')
    act(() => { vi.advanceTimersByTime(300) })
    expect(result.current).toBe('b')
  })

  it('resets the timer when the value changes before the delay elapses', () => {
    const { result, rerender } = renderHook(
      ({ value }) => useDebouncedValue(value, 300),
      { initialProps: { value: 'a' } },
    )
    rerender({ value: 'b' })
    act(() => { vi.advanceTimersByTime(200) })
    rerender({ value: 'c' })
    act(() => { vi.advanceTimersByTime(200) })
    expect(result.current).toBe('a')
    act(() => { vi.advanceTimersByTime(100) })
    expect(result.current).toBe('c')
  })
})

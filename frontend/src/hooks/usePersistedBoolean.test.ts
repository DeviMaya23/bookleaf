import { renderHook, act } from '@testing-library/react'
import { vi, describe, it, expect, beforeEach } from 'vitest'
import { usePersistedBoolean } from './usePersistedBoolean'

describe('usePersistedBoolean', () => {
  beforeEach(() => { localStorage.clear() })

  it('initialises from localStorage when a stored value exists', () => {
    localStorage.setItem('my-key', 'true')
    const { result } = renderHook(() => usePersistedBoolean('my-key', false))
    expect(result.current[0]).toBe(true)
  })

  it('falls back to defaultValue when nothing is stored', () => {
    const { result } = renderHook(() => usePersistedBoolean('my-key', true))
    expect(result.current[0]).toBe(true)
  })

  it('updates state and persists to localStorage when set is called', () => {
    const { result } = renderHook(() => usePersistedBoolean('my-key', false))
    act(() => { result.current[1](true) })
    expect(result.current[0]).toBe(true)
    expect(localStorage.getItem('my-key')).toBe('true')
  })

  it('initialises from defaultValue when localStorage.getItem throws', () => {
    vi.spyOn(Storage.prototype, 'getItem').mockImplementation(() => { throw new Error('blocked') })
    const { result } = renderHook(() => usePersistedBoolean('my-key', true))
    expect(result.current[0]).toBe(true)
    vi.restoreAllMocks()
  })

  it('updates state even when localStorage.setItem throws', () => {
    vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => { throw new Error('blocked') })
    const { result } = renderHook(() => usePersistedBoolean('my-key', false))
    act(() => { result.current[1](true) })
    expect(result.current[0]).toBe(true)
    vi.restoreAllMocks()
  })
})

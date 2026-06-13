import { describe, it, expect, beforeEach } from 'vitest'
import { act, renderHook } from '@testing-library/react'
import { ThemeProvider, useTheme } from './useTheme'

describe('useTheme', () => {
  beforeEach(() => {
    localStorage.clear()
    document.documentElement.removeAttribute('data-theme')
  })

  it('defaults to the warm theme when no preference is stored', () => {
    const { result } = renderHook(() => useTheme(), { wrapper: ThemeProvider })

    expect(result.current.theme).toBe('warm')
    expect(document.documentElement.dataset.theme).toBe('warm')
  })

  it('restores a stored theme preference on load', () => {
    localStorage.setItem('bookleaf-theme', 'warm')

    const { result } = renderHook(() => useTheme(), { wrapper: ThemeProvider })

    expect(result.current.theme).toBe('warm')
    expect(document.documentElement.dataset.theme).toBe('warm')
  })

  it('setTheme updates the data-theme attribute and localStorage', () => {
    const { result } = renderHook(() => useTheme(), { wrapper: ThemeProvider })

    act(() => result.current.setTheme('warm'))

    expect(document.documentElement.dataset.theme).toBe('warm')
    expect(localStorage.getItem('bookleaf-theme')).toBe('warm')
  })

  it('throws when used outside of ThemeProvider', () => {
    expect(() => renderHook(() => useTheme())).toThrow('useTheme must be used within a ThemeProvider')
  })
})

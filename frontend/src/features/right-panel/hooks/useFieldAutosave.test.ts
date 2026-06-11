import { describe, it, expect, vi } from 'vitest'
import { act, renderHook } from '@testing-library/react'
import { useFieldAutosave } from './useFieldAutosave'

describe('useFieldAutosave', () => {
  it('saves the new value on blur when it changed', () => {
    const onSave = vi.fn()
    const { result } = renderHook(() => useFieldAutosave<string>('original', onSave))

    act(() => result.current.onChange('edited'))
    act(() => result.current.onBlur())

    expect(onSave).toHaveBeenCalledWith('edited')
    expect(result.current.value).toBe('edited')
  })

  it('does not save on blur when the value is unchanged', () => {
    const onSave = vi.fn()
    const { result } = renderHook(() => useFieldAutosave<string>('original', onSave))

    act(() => result.current.onBlur())

    expect(onSave).not.toHaveBeenCalled()
  })

  it('reverts to the last saved value and does not save when isEmpty reports empty', () => {
    const onSave = vi.fn()
    const { result } = renderHook(() =>
      useFieldAutosave<string>('original', onSave, { isEmpty: (v) => v.trim() === '' }),
    )

    act(() => result.current.onChange('   '))
    act(() => result.current.onBlur())

    expect(onSave).not.toHaveBeenCalled()
    expect(result.current.value).toBe('original')
  })

  it('resets the local value when the source value changes', () => {
    const onSave = vi.fn()
    const { result, rerender } = renderHook(({ v }) => useFieldAutosave(v, onSave), {
      initialProps: { v: 'first' },
    })

    act(() => result.current.onChange('dirty edit'))
    rerender({ v: 'second' })

    expect(result.current.value).toBe('second')
    // blur after reset should compare against the new source, not save unchanged
    act(() => result.current.onBlur())
    expect(onSave).not.toHaveBeenCalled()
  })
})

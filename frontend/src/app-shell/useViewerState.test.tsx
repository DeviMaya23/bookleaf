import { renderHook, act } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import { useViewerState } from './useViewerState'
import type { Image } from '@/lib/images'

function makeImage(id: string): Image {
  return {
    id,
    title: `Image ${id}`,
    description: null,
    mime_type: 'image/jpeg',
    source_url: null,
    folder_ids: [],
    thumbnail_url: null,
    width: null,
    height: null,
    file_size: null,
    tags: [],
    position: null,
    created_at: '',
    updated_at: '',
  }
}

describe('useViewerState', () => {
  it('handleImageSelect opens lightbox on coarse pointer', () => {
    const { result } = renderHook(() =>
      useViewerState({ isCoarsePointer: true, viewKey: 'all' }),
    )
    act(() => result.current.handleImageSelect(makeImage('a')))
    expect(result.current.lightboxImage?.id).toBe('a')
    expect(result.current.selectedImage).toBeNull()
  })

  it('handleImageSelect sets selectedImage with no autoFocus on fine pointer', () => {
    const { result } = renderHook(() =>
      useViewerState({ isCoarsePointer: false, viewKey: 'all' }),
    )
    act(() => result.current.selectImageAutoFocus(makeImage('a')))
    act(() => result.current.handleImageSelect(makeImage('b')))
    expect(result.current.selectedImage?.id).toBe('b')
    expect(result.current.autoFocusTitle).toBe(false)
    expect(result.current.lightboxImage).toBeNull()
  })

  it('handleImageDoubleClick sets viewerImage and selectedImage on fine pointer', () => {
    const { result } = renderHook(() =>
      useViewerState({ isCoarsePointer: false, viewKey: 'all' }),
    )
    act(() => result.current.handleImageDoubleClick(makeImage('a')))
    expect(result.current.viewerImage?.id).toBe('a')
    expect(result.current.selectedImage?.id).toBe('a')
  })

  it('handleImageDoubleClick is a no-op on coarse pointer', () => {
    const { result } = renderHook(() =>
      useViewerState({ isCoarsePointer: true, viewKey: 'all' }),
    )
    act(() => result.current.handleImageDoubleClick(makeImage('a')))
    expect(result.current.viewerImage).toBeNull()
  })

  it('handleImageDeleted clears selectedImage only when id matches', () => {
    const { result } = renderHook(() =>
      useViewerState({ isCoarsePointer: false, viewKey: 'all' }),
    )
    act(() => result.current.selectImageAutoFocus(makeImage('a')))
    act(() => result.current.handleImageDeleted('other'))
    expect(result.current.selectedImage?.id).toBe('a')
    act(() => result.current.handleImageDeleted('a'))
    expect(result.current.selectedImage).toBeNull()
    expect(result.current.autoFocusTitle).toBe(false)
  })

  it('handleImageDeleted clears viewerImage only when id matches', () => {
    const { result } = renderHook(() =>
      useViewerState({ isCoarsePointer: false, viewKey: 'all' }),
    )
    act(() => result.current.handleImageDoubleClick(makeImage('a')))
    act(() => result.current.handleImageDeleted('other'))
    expect(result.current.viewerImage?.id).toBe('a')
    act(() => result.current.handleImageDeleted('a'))
    expect(result.current.viewerImage).toBeNull()
  })

  it('resets all state on viewKey change', () => {
    const { result, rerender } = renderHook(
      ({ viewKey }) => useViewerState({ isCoarsePointer: false, viewKey }),
      { initialProps: { viewKey: 'all' } },
    )
    act(() => {
      result.current.selectImageAutoFocus(makeImage('a'))
    })
    act(() => {
      result.current.handleImageDoubleClick(makeImage('a'))
    })
    rerender({ viewKey: 'folder:123' })
    expect(result.current.selectedImage).toBeNull()
    expect(result.current.viewerImage).toBeNull()
    expect(result.current.lightboxImage).toBeNull()
    expect(result.current.autoFocusTitle).toBe(false)
  })

  it('selectImageAutoFocus sets selectedImage and autoFocusTitle', () => {
    const { result } = renderHook(() =>
      useViewerState({ isCoarsePointer: false, viewKey: 'all' }),
    )
    act(() => result.current.selectImageAutoFocus(makeImage('a')))
    expect(result.current.selectedImage?.id).toBe('a')
    expect(result.current.autoFocusTitle).toBe(true)
  })

  it('clearImageSelection clears selectedImage and autoFocusTitle', () => {
    const { result } = renderHook(() =>
      useViewerState({ isCoarsePointer: false, viewKey: 'all' }),
    )
    act(() => result.current.selectImageAutoFocus(makeImage('a')))
    act(() => result.current.clearImageSelection())
    expect(result.current.selectedImage).toBeNull()
    expect(result.current.autoFocusTitle).toBe(false)
  })
})

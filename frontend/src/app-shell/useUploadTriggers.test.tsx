import { renderHook, act } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import { useUploadTriggers } from './useUploadTriggers'

function makeImageFile(name = 'photo.jpg'): File {
  return new File(['data'], name, { type: 'image/jpeg' })
}

function firePaste(file: File | null) {
  const item = file
    ? { kind: 'file', type: file.type, getAsFile: () => file }
    : null
  const event = new Event('paste', { bubbles: true }) as ClipboardEvent
  Object.defineProperty(event, 'clipboardData', {
    value: { items: item ? [item] : [] },
  })
  document.dispatchEvent(event)
}

describe('useUploadTriggers', () => {
  it('paste event with an image file opens the upload modal', () => {
    const { result } = renderHook(() => useUploadTriggers())
    const file = makeImageFile()
    act(() => firePaste(file))
    expect(result.current.uploadOpen).toBe(true)
    expect(result.current.uploadInitialFile).toBe(file)
  })

  it('paste event is ignored when upload modal is already open', () => {
    const { result } = renderHook(() => useUploadTriggers())
    act(() => result.current.openUpload())
    expect(result.current.uploadOpen).toBe(true)
    act(() => firePaste(makeImageFile()))
    // uploadInitialFile should stay null since openUpload was called without a file
    expect(result.current.uploadInitialFile).toBeNull()
  })

  it('paste event is ignored when target is an INPUT element', () => {
    const { result } = renderHook(() => useUploadTriggers())
    const input = document.createElement('input')
    document.body.appendChild(input)
    const event = new Event('paste', { bubbles: true }) as ClipboardEvent
    Object.defineProperty(event, 'target', { value: input })
    Object.defineProperty(event, 'clipboardData', {
      value: { items: [{ kind: 'file', type: 'image/jpeg', getAsFile: () => makeImageFile() }] },
    })
    act(() => document.dispatchEvent(event))
    expect(result.current.uploadOpen).toBe(false)
    document.body.removeChild(input)
  })

  it('openBatchUpload sets batchInitialFiles and opens the batch modal', () => {
    const { result } = renderHook(() => useUploadTriggers())
    const files = [makeImageFile('a.jpg'), makeImageFile('b.jpg')]
    act(() => result.current.openBatchUpload(files))
    expect(result.current.batchUploadOpen).toBe(true)
    expect(result.current.batchInitialFiles).toBe(files)
  })

  it('closeUpload clears uploadInitialFile and closes the modal', () => {
    const { result } = renderHook(() => useUploadTriggers())
    const file = makeImageFile()
    act(() => result.current.openUpload(file))
    act(() => result.current.closeUpload())
    expect(result.current.uploadOpen).toBe(false)
    expect(result.current.uploadInitialFile).toBeNull()
  })
})

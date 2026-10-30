import { renderHook, act } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { vi, describe, it, expect, beforeEach } from 'vitest'
import { useFileDrop } from './useFileDrop'
import { handleFileAutoUpload } from './lib/dragHandlers'
import type { AutoUploadResult } from './lib/dragHandlers'
import type { Image } from '@/lib/images'

vi.mock('@kinde-oss/kinde-auth-react', () => ({
  useKindeAuth: () => ({ getToken: vi.fn().mockResolvedValue('token') }),
}))

vi.mock('./lib/dragHandlers', async (importOriginal) => ({
  ...(await importOriginal<typeof import('./lib/dragHandlers')>()),
  handleFileAutoUpload: vi.fn(),
}))

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn(), warning: vi.fn() },
}))

function makeWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return {
    queryClient,
    wrapper: ({ children }: { children: React.ReactNode }) => (
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    ),
  }
}

function makeImage(id = 'img-1'): Image {
  return {
    id,
    title: 'Test',
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

function makeDragEvent(files: File[]): React.DragEvent<HTMLElement> {
  return {
    preventDefault: vi.fn(),
    dataTransfer: {
      files,
      types: ['Files'],
    },
    currentTarget: document.createElement('div'),
    relatedTarget: null,
  } as unknown as React.DragEvent<HTMLElement>
}

describe('useFileDrop', () => {
  beforeEach(() => { vi.clearAllMocks() })

  it('single file drop auto-uploads and calls onAutoUploaded', async () => {
    const image = makeImage()
    vi.mocked(handleFileAutoUpload).mockResolvedValue({ image, duplicates: [] } as AutoUploadResult)
    const onAutoUploaded = vi.fn()
    const { wrapper } = makeWrapper()
    const { result } = renderHook(
      () => useFileDrop({ folderId: null, onAutoUploaded, onBatchFilesDropped: vi.fn() }),
      { wrapper },
    )
    const file = new File(['data'], 'photo.jpg', { type: 'image/jpeg' })
    await act(async () => {
      await result.current.dragHandlers.onDrop(makeDragEvent([file]))
    })
    expect(handleFileAutoUpload).toHaveBeenCalledWith(expect.any(Function), file, null)
    expect(onAutoUploaded).toHaveBeenCalledWith(image)
    expect(result.current.isAutoUploading).toBe(false)
  })

  it('multi-file drop calls onBatchFilesDropped without uploading', async () => {
    const onBatchFilesDropped = vi.fn()
    const { wrapper } = makeWrapper()
    const { result } = renderHook(
      () => useFileDrop({ folderId: null, onAutoUploaded: vi.fn(), onBatchFilesDropped }),
      { wrapper },
    )
    const files = [
      new File(['a'], 'a.jpg', { type: 'image/jpeg' }),
      new File(['b'], 'b.jpg', { type: 'image/jpeg' }),
    ]
    await act(async () => {
      await result.current.dragHandlers.onDrop(makeDragEvent(files))
    })
    expect(onBatchFilesDropped).toHaveBeenCalledWith(files)
    expect(handleFileAutoUpload).not.toHaveBeenCalled()
  })

  it('dragOver with Files type sets isFileDragOver', () => {
    const { wrapper } = makeWrapper()
    const { result } = renderHook(
      () => useFileDrop({ folderId: null, onAutoUploaded: vi.fn(), onBatchFilesDropped: vi.fn() }),
      { wrapper },
    )
    const event = {
      preventDefault: vi.fn(),
      dataTransfer: { types: ['Files'] },
    } as unknown as React.DragEvent<HTMLElement>
    act(() => result.current.dragHandlers.onDragOver(event))
    expect(result.current.isFileDragOver).toBe(true)
  })

  it('dragLeave clears isFileDragOver when leaving the drop zone', () => {
    const { wrapper } = makeWrapper()
    const { result } = renderHook(
      () => useFileDrop({ folderId: null, onAutoUploaded: vi.fn(), onBatchFilesDropped: vi.fn() }),
      { wrapper },
    )
    const container = document.createElement('div')
    const event = {
      preventDefault: vi.fn(),
      dataTransfer: { types: ['Files'] },
      currentTarget: container,
      relatedTarget: null,
    } as unknown as React.DragEvent<HTMLElement>
    act(() => result.current.dragHandlers.onDragOver(event))
    act(() => result.current.dragHandlers.onDragLeave(event))
    expect(result.current.isFileDragOver).toBe(false)
  })
})

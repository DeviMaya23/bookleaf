import { vi, describe, it, expect, beforeEach } from 'vitest'

vi.mock('./images', () => ({
  moveImageFolder: vi.fn(),
  initiateUpload: vi.fn(),
  putToR2: vi.fn(),
  completeUpload: vi.fn(),
  getImage: vi.fn(),
}))

vi.mock('./folders', () => ({
  getFolderSubtreeIds: vi.fn(),
  moveFolder: vi.fn(),
}))

import { handleImageDrop, handleFolderDrop, handleFileAutoUpload } from './dragHandlers'
import { moveImageFolder, initiateUpload, putToR2, completeUpload, getImage } from './images'
import { moveFolder, getFolderSubtreeIds } from './folders'
import type { Folder } from './folders'

const getToken = vi.fn().mockResolvedValue('token')

function makeFolder(id: string, parentId: string | null = null): Folder {
  return {
    id,
    name: `Folder ${id}`,
    description: null,
    parent_id: parentId,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
  }
}

describe('handleImageDrop', () => {
  beforeEach(() => vi.clearAllMocks())

  it('calls moveImageFolder with target folder_id on success', async () => {
    vi.mocked(moveImageFolder).mockResolvedValueOnce(undefined)

    const result = await handleImageDrop(
      getToken,
      { type: 'image', imageId: 'img-1', currentFolderId: 'folder-a' },
      { type: 'folder', folderId: 'folder-b' },
    )

    expect(result).toBe('moved')
    expect(moveImageFolder).toHaveBeenCalledWith(getToken, 'img-1', 'folder-a', 'folder-b')
  })

  it('returns noop when image is already in the target folder', async () => {
    const result = await handleImageDrop(
      getToken,
      { type: 'image', imageId: 'img-1', currentFolderId: 'folder-a' },
      { type: 'folder', folderId: 'folder-a' },
    )

    expect(result).toBe('noop')
    expect(moveImageFolder).not.toHaveBeenCalled()
  })

  it('calls moveImageFolder with null target when dropped on unsorted', async () => {
    vi.mocked(moveImageFolder).mockResolvedValueOnce(undefined)

    const result = await handleImageDrop(
      getToken,
      { type: 'image', imageId: 'img-1', currentFolderId: 'folder-a' },
      { type: 'unsorted' },
    )

    expect(result).toBe('moved')
    expect(moveImageFolder).toHaveBeenCalledWith(getToken, 'img-1', 'folder-a', null)
  })

  it('throws when moveImageFolder fails', async () => {
    vi.mocked(moveImageFolder).mockRejectedValueOnce(new Error('network error'))

    await expect(
      handleImageDrop(
        getToken,
        { type: 'image', imageId: 'img-1', currentFolderId: null },
        { type: 'folder', folderId: 'folder-b' },
      ),
    ).rejects.toThrow('network error')
  })
})

describe('handleFolderDrop', () => {
  beforeEach(() => vi.clearAllMocks())

  it('calls moveFolder with target parent_id on success', async () => {
    vi.mocked(getFolderSubtreeIds).mockReturnValueOnce(new Set(['folder-a']))
    vi.mocked(moveFolder).mockResolvedValueOnce({} as never)

    const folders = [makeFolder('folder-a'), makeFolder('folder-b')]
    const result = await handleFolderDrop(
      getToken,
      { type: 'folder', folderId: 'folder-a', name: 'Folder A', parentId: null },
      { type: 'folder', folderId: 'folder-b' },
      folders,
    )

    expect(result).toBe('moved')
    expect(moveFolder).toHaveBeenCalledWith(getToken, 'folder-a', 'Folder A', 'folder-b')
  })

  it('returns noop when folder is already a child of target', async () => {
    vi.mocked(getFolderSubtreeIds).mockReturnValueOnce(new Set(['folder-a']))

    const folders = [makeFolder('folder-a', 'folder-b'), makeFolder('folder-b')]
    const result = await handleFolderDrop(
      getToken,
      { type: 'folder', folderId: 'folder-a', name: 'Folder A', parentId: 'folder-b' },
      { type: 'folder', folderId: 'folder-b' },
      folders,
    )

    expect(result).toBe('noop')
    expect(moveFolder).not.toHaveBeenCalled()
  })

  it('returns circular when target is in dragged folder subtree', async () => {
    vi.mocked(getFolderSubtreeIds).mockReturnValueOnce(new Set(['folder-a', 'folder-b']))

    const folders = [makeFolder('folder-a'), makeFolder('folder-b', 'folder-a')]
    const result = await handleFolderDrop(
      getToken,
      { type: 'folder', folderId: 'folder-a', name: 'Folder A', parentId: null },
      { type: 'folder', folderId: 'folder-b' },
      folders,
    )

    expect(result).toBe('circular')
    expect(moveFolder).not.toHaveBeenCalled()
  })

  it('calls moveFolder with null parent_id when dropped on root zone', async () => {
    vi.mocked(getFolderSubtreeIds).mockReturnValueOnce(new Set(['folder-a']))
    vi.mocked(moveFolder).mockResolvedValueOnce({} as never)

    const folders = [makeFolder('folder-a', 'folder-b')]
    const result = await handleFolderDrop(
      getToken,
      { type: 'folder', folderId: 'folder-a', name: 'Folder A', parentId: 'folder-b' },
      { type: 'root' },
      folders,
    )

    expect(result).toBe('moved')
    expect(moveFolder).toHaveBeenCalledWith(getToken, 'folder-a', 'Folder A', null)
  })

  it('throws when moveFolder fails', async () => {
    vi.mocked(getFolderSubtreeIds).mockReturnValueOnce(new Set(['folder-a']))
    vi.mocked(moveFolder).mockRejectedValueOnce(new Error('server error'))

    const folders = [makeFolder('folder-a')]
    await expect(
      handleFolderDrop(
        getToken,
        { type: 'folder', folderId: 'folder-a', name: 'Folder A', parentId: null },
        { type: 'folder', folderId: 'folder-b' },
        folders,
      ),
    ).rejects.toThrow('server error')
  })
})

describe('handleFileAutoUpload', () => {
  beforeEach(() => vi.clearAllMocks())

  function makeFile(name: string, type: string): File {
    return new File(['bytes'], name, { type })
  }

  it('completes 3-step upload and returns full Image on success', async () => {
    const imageData = { id: 'img-new', title: 'sunset', folder_ids: ['folder-1'] }
    vi.mocked(initiateUpload).mockResolvedValueOnce({ id: 'upload-1', upload_url: 'https://r2.example.com/upload', r2_path: 'path' })
    vi.mocked(putToR2).mockResolvedValueOnce(undefined)
    vi.mocked(completeUpload).mockResolvedValueOnce({ image_id: 'img-new', suggested_folder_name: null })
    vi.mocked(getImage).mockResolvedValueOnce(imageData as never)

    const file = makeFile('sunset.jpg', 'image/jpeg')
    const result = await handleFileAutoUpload(getToken, file, 'folder-1')

    expect(initiateUpload).toHaveBeenCalledWith(getToken, { title: 'sunset', mimeType: 'image/jpeg', folderId: 'folder-1' })
    expect(putToR2).toHaveBeenCalledWith('https://r2.example.com/upload', file)
    expect(completeUpload).toHaveBeenCalledWith(getToken, 'upload-1')
    expect(getImage).toHaveBeenCalledWith(getToken, 'img-new')
    expect(result).toBe(imageData)
  })

  it('throws unsupported_type error for invalid file type', async () => {
    const file = makeFile('doc.pdf', 'application/pdf')

    await expect(handleFileAutoUpload(getToken, file, null)).rejects.toThrow('unsupported_type')
    expect(initiateUpload).not.toHaveBeenCalled()
  })

  it('throws when initiateUpload fails', async () => {
    vi.mocked(initiateUpload).mockRejectedValueOnce(new Error('server error'))

    const file = makeFile('photo.png', 'image/png')
    await expect(handleFileAutoUpload(getToken, file, null)).rejects.toThrow('server error')
  })
})

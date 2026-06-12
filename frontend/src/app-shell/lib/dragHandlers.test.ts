import { vi, describe, it, expect, beforeEach } from 'vitest'

vi.mock('@/lib/images', () => ({
  moveImageFolder: vi.fn(),
  getImage: vi.fn(),
}))

vi.mock('@/lib/folders', () => ({
  getFolderSubtreeIds: vi.fn(),
  updateFolder: vi.fn(),
}))

vi.mock('@/lib/upload', () => ({
  validateImageFile: vi.fn(),
  uploadImageFile: vi.fn(),
}))

import { handleImageDrop, handleFolderDrop, handleFileAutoUpload } from './dragHandlers'
import { moveImageFolder, getImage } from '@/lib/images'
import { updateFolder, getFolderSubtreeIds } from '@/lib/folders'
import { validateImageFile, uploadImageFile } from '@/lib/upload'
import type { Folder } from '@/lib/folders'

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

  it('calls updateFolder with target parent_id on success', async () => {
    vi.mocked(getFolderSubtreeIds).mockReturnValueOnce(new Set(['folder-a']))
    vi.mocked(updateFolder).mockResolvedValueOnce({} as never)

    const folders = [makeFolder('folder-a'), makeFolder('folder-b')]
    const result = await handleFolderDrop(
      getToken,
      { type: 'folder', folderId: 'folder-a', name: 'Folder A', parentId: null },
      { type: 'folder', folderId: 'folder-b' },
      folders,
    )

    expect(result).toBe('moved')
    expect(updateFolder).toHaveBeenCalledWith(getToken, 'folder-a', { parent_id: 'folder-b' })
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
    expect(updateFolder).not.toHaveBeenCalled()
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
    expect(updateFolder).not.toHaveBeenCalled()
  })

  it('calls updateFolder with null parent_id when dropped on root zone', async () => {
    vi.mocked(getFolderSubtreeIds).mockReturnValueOnce(new Set(['folder-a']))
    vi.mocked(updateFolder).mockResolvedValueOnce({} as never)

    const folders = [makeFolder('folder-a', 'folder-b')]
    const result = await handleFolderDrop(
      getToken,
      { type: 'folder', folderId: 'folder-a', name: 'Folder A', parentId: 'folder-b' },
      { type: 'root' },
      folders,
    )

    expect(result).toBe('moved')
    expect(updateFolder).toHaveBeenCalledWith(getToken, 'folder-a', { parent_id: null })
  })

  it('throws when updateFolder fails', async () => {
    vi.mocked(getFolderSubtreeIds).mockReturnValueOnce(new Set(['folder-a']))
    vi.mocked(updateFolder).mockRejectedValueOnce(new Error('server error'))

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

  it('returns full ImageDetail via getImage on success', async () => {
    const imageData = { id: 'img-new', title: 'sunset', folder_ids: ['folder-1'] }
    vi.mocked(validateImageFile).mockReturnValueOnce(null)
    vi.mocked(uploadImageFile).mockResolvedValueOnce({ image_id: 'img-new', suggested_folder_name: null })
    vi.mocked(getImage).mockResolvedValueOnce(imageData as never)

    const file = makeFile('sunset.jpg', 'image/jpeg')
    const result = await handleFileAutoUpload(getToken, file, 'folder-1')

    expect(uploadImageFile).toHaveBeenCalledWith(getToken, { file, folderId: 'folder-1' })
    expect(getImage).toHaveBeenCalledWith(getToken, 'img-new')
    expect(result).toBe(imageData)
  })

  it('rejects unsupported type without calling uploadImageFile', async () => {
    vi.mocked(validateImageFile).mockReturnValueOnce('unsupported_type')
    const file = makeFile('doc.pdf', 'application/pdf')

    await expect(handleFileAutoUpload(getToken, file, null)).rejects.toThrow('unsupported_type')
    expect(uploadImageFile).not.toHaveBeenCalled()
  })
})

import { moveImageFolder, getImage } from '@/lib/images'
import { validateImageFile, uploadImageFile } from '@/lib/upload'
import { updateFolder } from '@/lib/folders'
import { getFolderSubtreeIds } from '@/lib/folders'
import type { Folder } from '@/lib/folders'
import type { Image } from '@/lib/images'

type GetToken = () => Promise<string | undefined>

export interface ImageDragData {
  type: 'image'
  imageId: string
  currentFolderId: string | null
  thumbnailUrl: string | null
}

export interface FolderDragData {
  type: 'folder'
  folderId: string
  name: string
  parentId: string | null
}

interface FolderDropData {
  type: 'folder'
  folderId: string
}

interface UnsortedDropData {
  type: 'unsorted'
}

interface RootDropData {
  type: 'root'
}

export type DropData = FolderDropData | UnsortedDropData | RootDropData

export function isImageDragData(data: unknown): data is ImageDragData {
  return (
    typeof data === 'object' &&
    data !== null &&
    (data as Record<string, unknown>).type === 'image' &&
    typeof (data as Record<string, unknown>).imageId === 'string'
  )
}

export function isFolderDragData(data: unknown): data is FolderDragData {
  return (
    typeof data === 'object' &&
    data !== null &&
    (data as Record<string, unknown>).type === 'folder' &&
    typeof (data as Record<string, unknown>).folderId === 'string' &&
    typeof (data as Record<string, unknown>).name === 'string'
  )
}

export function isDropData(data: unknown): data is DropData {
  if (typeof data !== 'object' || data === null) return false
  const type = (data as Record<string, unknown>).type
  if (type === 'unsorted' || type === 'root') return true
  if (type === 'folder') return typeof (data as Record<string, unknown>).folderId === 'string'
  return false
}

export async function handleImageDrop(
  getToken: GetToken,
  drag: ImageDragData,
  drop: DropData,
): Promise<'moved' | 'noop'> {
  if (drop.type === 'folder') {
    if (drag.currentFolderId === drop.folderId) return 'noop'
    await moveImageFolder(getToken, drag.imageId, drag.currentFolderId, drop.folderId)
    return 'moved'
  }
  if (drop.type === 'unsorted') {
    if (drag.currentFolderId === null) return 'noop'
    await moveImageFolder(getToken, drag.imageId, drag.currentFolderId, null)
    return 'moved'
  }
  return 'noop'
}

export async function handleFolderDrop(
  getToken: GetToken,
  drag: FolderDragData,
  drop: DropData,
  folders: Folder[],
): Promise<'moved' | 'noop' | 'circular'> {
  const subtreeIds = getFolderSubtreeIds(folders, drag.folderId)

  if (drop.type === 'folder') {
    if (subtreeIds.has(drop.folderId)) return 'circular'
    if (drag.parentId === drop.folderId) return 'noop'
    await updateFolder(getToken, drag.folderId, { parent_id: drop.folderId })
    return 'moved'
  }
  if (drop.type === 'root') {
    if (drag.parentId === null) return 'noop'
    await updateFolder(getToken, drag.folderId, { parent_id: null })
    return 'moved'
  }
  return 'noop'
}

export interface AutoUploadResult {
  image: Image
  duplicates: Array<{ id: string; title: string; thumbnail_path: string | null }>
}

export async function handleFileAutoUpload(
  getToken: GetToken,
  file: File,
  folderId: string | null,
): Promise<AutoUploadResult> {
  const err = validateImageFile(file)
  if (err) {
    throw new Error(err)
  }

  const result = await uploadImageFile(getToken, { file, folderId })
  const image = await getImage(getToken, result.image_id)
  return { image, duplicates: result.duplicates ?? [] }
}

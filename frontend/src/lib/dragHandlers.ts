import { updateImage, initiateUpload, putToR2, completeUpload, getImage } from './images'
import { moveFolder } from './folders'
import { getFolderSubtreeIds } from '@/components/FolderSidebar'
import type { Folder } from './folders'
import type { Image } from './images'

type GetToken = () => Promise<string | undefined>

interface ImageDragData {
  type: 'image'
  imageId: string
  currentFolderId: string | null
}

interface FolderDragData {
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

type DropData = FolderDropData | UnsortedDropData | RootDropData

export async function handleImageDrop(
  getToken: GetToken,
  drag: ImageDragData,
  drop: DropData,
): Promise<'moved' | 'noop'> {
  if (drop.type === 'folder') {
    if (drag.currentFolderId === drop.folderId) return 'noop'
    await updateImage(getToken, drag.imageId, { folder_id: drop.folderId })
    return 'moved'
  }
  if (drop.type === 'unsorted') {
    if (drag.currentFolderId === null) return 'noop'
    await updateImage(getToken, drag.imageId, { folder_id: null })
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
    await moveFolder(getToken, drag.folderId, drag.name, drop.folderId)
    return 'moved'
  }
  if (drop.type === 'root') {
    if (drag.parentId === null) return 'noop'
    await moveFolder(getToken, drag.folderId, drag.name, null)
    return 'moved'
  }
  return 'noop'
}

const ACCEPTED_TYPES = ['image/jpeg', 'image/png', 'image/gif', 'image/webp']

function fileBaseName(name: string): string {
  return name.replace(/\.[^.]+$/, '')
}

export async function handleFileAutoUpload(
  getToken: GetToken,
  file: File,
  folderId: string | null,
): Promise<Image> {
  if (!ACCEPTED_TYPES.includes(file.type)) {
    throw new Error('unsupported_type')
  }
  const title = fileBaseName(file.name)
  const initiated = await initiateUpload(getToken, {
    title,
    mimeType: file.type,
    folderId: folderId ?? undefined,
  })
  await putToR2(initiated.upload_url, file)
  const result = await completeUpload(getToken, initiated.id)
  return getImage(getToken, result.image_id)
}

import { moveImageFolder, initiateUpload, putToR2, completeUpload, getImage } from './images'
import { generateThumbnail, convertHeicToJpeg } from './thumbnail'
import { isSafari } from './browser'
import { moveFolder } from './folders'
import { getFolderSubtreeIds } from './folders'
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

const ACCEPTED_TYPES = ['image/jpeg', 'image/png', 'image/gif', 'image/webp', 'image/avif', 'image/heic']

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
  if (file.type === 'image/heic' && !isSafari()) {
    throw new Error('heic_safari_only')
  }

  let uploadBlob: Blob | File = file
  let mimeType = file.type
  if (file.type === 'image/heic') {
    uploadBlob = await convertHeicToJpeg(file)
    mimeType = 'image/jpeg'
  }

  const title = fileBaseName(file.name)
  const initiated = await initiateUpload(getToken, {
    title,
    mimeType,
    folderId: folderId ?? undefined,
  })

  const thumbnail = await generateThumbnail(uploadBlob)
  await Promise.all([
    putToR2(initiated.upload_url, uploadBlob),
    putToR2(initiated.thumbnail_upload_url, thumbnail),
  ])

  const result = await completeUpload(getToken, initiated.id)
  return getImage(getToken, result.image_id)
}

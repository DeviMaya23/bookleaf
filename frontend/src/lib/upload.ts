import { initiateUpload, putToR2, completeUpload } from '@/lib/images'
import type { CompleteUploadResult } from '@/lib/images'
import { generateThumbnail, convertHeicToJpeg } from '@/lib/thumbnail'
import { isSafari } from '@/lib/browser'

type GetToken = () => Promise<string | undefined>

const ACCEPTED_TYPES = ['image/jpeg', 'image/png', 'image/gif', 'image/webp', 'image/avif', 'image/heic']

export type FileValidationError = 'unsupported_type' | 'heic_safari_only'

export function validateImageFile(file: File): FileValidationError | null {
  if (!ACCEPTED_TYPES.includes(file.type)) {
    return 'unsupported_type'
  }
  if (file.type === 'image/heic' && !isSafari()) {
    return 'heic_safari_only'
  }
  return null
}

export function fileBaseName(name: string): string {
  return name.replace(/\.[^.]+$/, '')
}

export interface UploadImageFileParams {
  file: File
  folderId: string | null
  title?: string
  description?: string
  sourceUrl?: string
}

export async function uploadImageFile(
  getToken: GetToken,
  params: UploadImageFileParams,
): Promise<CompleteUploadResult> {
  const { file, folderId, description, sourceUrl } = params
  const title = params.title ?? fileBaseName(file.name)

  let uploadBlob: Blob | File = file
  let mimeType = file.type
  if (file.type === 'image/heic') {
    uploadBlob = await convertHeicToJpeg(file)
    mimeType = 'image/jpeg'
  }

  const initiated = await initiateUpload(getToken, {
    title,
    mimeType,
    folderId: folderId ?? undefined,
    description,
    sourceUrl,
  })

  const { blob: thumbnail, width, height } = await generateThumbnail(uploadBlob)
  await Promise.all([
    putToR2(initiated.upload_url, uploadBlob),
    putToR2(initiated.thumbnail_upload_url, thumbnail),
  ])

  return completeUpload(getToken, initiated.id, { width, height, fileSize: uploadBlob.size })
}

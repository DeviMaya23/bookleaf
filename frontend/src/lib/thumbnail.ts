import { computePHash } from '@/lib/phash'

const THUMBNAIL_MAX_PX = 600
const THUMBNAIL_QUALITY = 0.9
const HEIC_QUALITY = 0.93

export interface ThumbnailResult {
  blob: Blob
  width: number
  height: number
  phash: string
}

export async function generateThumbnail(source: Blob): Promise<ThumbnailResult> {
  const bitmap = await createImageBitmap(source)

  const { width, height } = bitmap
  const scale = Math.min(1, THUMBNAIL_MAX_PX / Math.max(width, height))
  const tw = Math.round(width * scale)
  const th = Math.round(height * scale)

  const canvas = new OffscreenCanvas(tw, th)
  const ctx = canvas.getContext('2d')!
  ctx.drawImage(bitmap, 0, 0, tw, th)
  bitmap.close()

  const phash = computePHash(canvas)
  const blob = await canvas.convertToBlob({ type: 'image/jpeg', quality: THUMBNAIL_QUALITY })
  return { blob, width, height, phash }
}

export async function convertHeicToJpeg(file: File): Promise<Blob> {
  const bitmap = await createImageBitmap(file)

  const canvas = new OffscreenCanvas(bitmap.width, bitmap.height)
  const ctx = canvas.getContext('2d')!
  ctx.drawImage(bitmap, 0, 0)
  bitmap.close()

  return canvas.convertToBlob({ type: 'image/jpeg', quality: HEIC_QUALITY })
}

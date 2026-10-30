import type { MasonryImage } from '@/features/gallery/components/MasonryLayout'
import type { SharedImage } from '@/lib/share'

export function toGalleryImage(image: SharedImage, index: number): MasonryImage {
  return {
    id: String(index),
    title: image.title,
    thumbnail_url: image.thumbnail_url,
    width: image.width,
    height: image.height,
  }
}

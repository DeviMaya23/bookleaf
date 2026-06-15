import type { Image } from '@/lib/images'
import type { SharedImage } from '@/lib/share'

export function toGalleryImage(image: SharedImage, index: number): Image {
  return {
    id: String(index),
    title: image.title,
    description: null,
    mime_type: '',
    source_url: null,
    folder_ids: [],
    thumbnail_url: image.thumbnail_url,
    width: image.width,
    height: image.height,
    file_size: null,
    tags: [],
    position: null,
    created_at: '',
    updated_at: '',
  }
}

import { useEffect, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useKindeAuth } from '@kinde-oss/kinde-auth-react'
import { X } from 'lucide-react'
import { getImage } from '@/lib/images'
import type { Image } from '@/lib/images'

interface ImageLightboxProps {
  image: Image
  onClose: () => void
}

export default function ImageLightbox({ image, onClose }: ImageLightboxProps) {
  const { getToken } = useKindeAuth()

  const { data: imageDetail } = useQuery({
    queryKey: ['image', image.id],
    queryFn: () => getImage(getToken, image.id),
  })

  const displaySrc = imageDetail?.image_url

  const [loaded, setLoaded] = useState(false)
  useEffect(() => {
    setLoaded(false)
  }, [displaySrc, image.id])

  return (
    <div className="fixed inset-0 z-50 bg-black flex items-center justify-center">
      <button
        onClick={onClose}
        aria-label="Close"
        className="absolute top-4 right-4 w-9 h-9 rounded-full bg-black/40 text-white flex items-center justify-center"
      >
        <X className="w-5 h-5" />
      </button>

      {image.thumbnail_url && !loaded && (
        <img
          data-testid="lightbox-thumbnail"
          src={image.thumbnail_url}
          alt={image.title}
          draggable={false}
          className="max-w-full max-h-full object-contain"
        />
      )}
      {displaySrc && (
        <img
          data-testid="lightbox-full-image"
          src={displaySrc}
          alt={image.title}
          draggable={false}
          onLoad={() => setLoaded(true)}
          className="max-w-full max-h-full object-contain"
          style={{ position: loaded ? 'static' : 'absolute', opacity: loaded ? 1 : 0 }}
        />
      )}
    </div>
  )
}

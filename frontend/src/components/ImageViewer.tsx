import { useEffect } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useKindeAuth } from '@kinde-oss/kinde-auth-react'
import { ChevronLeft, FlipHorizontal, RotateCw } from 'lucide-react'
import { getImage } from '@/lib/images'
import type { Image } from '@/lib/images'

interface ImageViewerProps {
  image: Image
  onClose: () => void
}

export default function ImageViewer({ image, onClose }: ImageViewerProps) {
  const { getToken } = useKindeAuth()

  const { data: imageDetail } = useQuery({
    queryKey: ['image', image.id],
    queryFn: () => getImage(getToken, image.id),
  })

  useEffect(() => {
    function handleKeyDown(e: KeyboardEvent) {
      if (e.key === 'Escape') onClose()
    }
    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [onClose])

  const displaySrc = imageDetail?.image_url ?? image.thumbnail_url

  return (
    <div className="flex flex-col w-full h-full">
      <div className="flex items-center h-11 px-2 gap-2 flex-shrink-0 border-b">
        <button onClick={onClose} aria-label="Back" className="p-1.5 rounded hover:bg-muted transition-colors">
          <ChevronLeft className="w-5 h-5" />
        </button>
        <input type="range" min={5} max={800} defaultValue={100} className="w-32" />
        <span className="text-sm tabular-nums">100%</span>
        <div className="h-4 w-px bg-border mx-1" />
        <button aria-label="Flip horizontal" className="p-1.5 rounded hover:bg-muted transition-colors cursor-pointer">
          <FlipHorizontal className="w-4 h-4" />
        </button>
        <button aria-label="Rotate 90° clockwise" className="p-1.5 rounded hover:bg-muted transition-colors cursor-pointer">
          <RotateCw className="w-4 h-4" />
        </button>
        <button className="text-sm px-2 py-1 rounded hover:bg-muted transition-colors cursor-pointer">1:1</button>
        <div className="flex-1" />
      </div>
      <div className="flex-1 overflow-hidden relative flex items-center justify-center">
        {displaySrc && (
          <img
            src={displaySrc}
            alt={image.title}
            className="max-w-full max-h-full object-contain"
          />
        )}
        <div className="absolute bottom-4 left-1/2 -translate-x-1/2 bg-black/50 text-white text-xs px-3 py-1 rounded-full whitespace-nowrap">
          {`${image.title} · ${image.width} × ${image.height}`}
        </div>
      </div>
    </div>
  )
}

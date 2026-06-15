import { useEffect } from 'react'
import { X, ChevronLeft, ChevronRight } from 'lucide-react'
import { Button } from '@/components/ui/button'
import type { SharedImage } from '@/lib/share'

interface LightboxProps {
  images: SharedImage[]
  index: number
  onClose: () => void
  onNavigate: (dir: 1 | -1) => void
}

export default function Lightbox({ images, index, onClose, onNavigate }: LightboxProps) {
  const hasPrev = index > 0
  const hasNext = index < images.length - 1
  const image = images[index]

  useEffect(() => {
    function handleKeyDown(e: KeyboardEvent) {
      if (e.key === 'Escape') onClose()
      else if (e.key === 'ArrowLeft' && hasPrev) onNavigate(-1)
      else if (e.key === 'ArrowRight' && hasNext) onNavigate(1)
    }
    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [onClose, onNavigate, hasPrev, hasNext])

  return (
    <div
      className="fixed inset-0 z-50 flex flex-col items-center justify-center bg-black/90"
      onClick={onClose}
    >
      <Button
        variant="secondary"
        size="icon-sm"
        className="absolute top-4 right-4 rounded-full"
        aria-label="Close"
        onClick={(e) => { e.stopPropagation(); onClose() }}
      >
        <X className="w-4 h-4" />
      </Button>

      {hasPrev && (
        <Button
          variant="secondary"
          size="icon-sm"
          className="absolute left-4 top-1/2 -translate-y-1/2 rounded-full"
          aria-label="Previous image"
          onClick={(e) => { e.stopPropagation(); onNavigate(-1) }}
        >
          <ChevronLeft className="w-4 h-4" />
        </Button>
      )}

      {hasNext && (
        <Button
          variant="secondary"
          size="icon-sm"
          className="absolute right-4 top-1/2 -translate-y-1/2 rounded-full"
          aria-label="Next image"
          onClick={(e) => { e.stopPropagation(); onNavigate(1) }}
        >
          <ChevronRight className="w-4 h-4" />
        </Button>
      )}

      <img
        src={image.full_res_url}
        alt={image.title}
        className="max-h-[85vh] max-w-[90vw] object-contain"
        onClick={(e) => e.stopPropagation()}
      />

      <div className="mt-4 flex flex-col items-center gap-1 text-white">
        <p className="text-sm">{image.title}</p>
        <p className="text-xs text-white/70">{index + 1} / {images.length}</p>
      </div>
    </div>
  )
}

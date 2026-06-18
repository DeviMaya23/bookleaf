import { useEffect, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useKindeAuth } from '@kinde-oss/kinde-auth-react'
import { ChevronLeft, FlipHorizontal, Focus, Loader2, RotateCw } from 'lucide-react'
import { Toggle } from '@/components/ui/toggle'
import { getImage } from '@/lib/images'
import type { Image } from '@/lib/images'
import { useImageTransform } from '../hooks/useImageTransform'

interface ImageViewerProps {
  image: Image
  onClose: () => void
  focusMode: boolean
  onToggleFocusMode: () => void
}

export default function ImageViewer({ image, onClose, focusMode, onToggleFocusMode }: ImageViewerProps) {
  const { getToken } = useKindeAuth()

  const { data: imageDetail } = useQuery({
    queryKey: ['image', image.id],
    queryFn: () => getImage(getToken, image.id),
  })

  const { containerRef, transform, zoom, setZoom, dragging, dragHandlers, toggleFlip, rotate, resetTo1to1 } = useImageTransform(image)

  // Esc to close
  useEffect(() => {
    function handleKeyDown(e: KeyboardEvent) {
      if (e.key === 'Escape') onClose()
    }
    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [onClose])

  const displaySrc = imageDetail?.image_url

  const [loaded, setLoaded] = useState(false)
  useEffect(() => {
    setLoaded(false)
  }, [displaySrc])

  return (
    <div className="flex flex-col w-full h-full">
      <div className="flex items-center h-11 px-2 gap-2 flex-shrink-0 border-b">
        <div className="hidden sm:flex">
          <Toggle
            aria-label="Focus mode"
            aria-pressed={focusMode}
            pressed={focusMode}
            onPressedChange={() => onToggleFocusMode()}
            className="aria-pressed:bg-secondary aria-pressed:text-secondary-foreground"
          >
            <Focus className="w-3.5 h-3.5" />
          </Toggle>
        </div>
        <button onClick={onClose} aria-label="Back" className="p-1.5 rounded hover:bg-muted transition-colors">
          <ChevronLeft className="w-5 h-5" />
        </button>
        <input
          type="range"
          min={5}
          max={800}
          value={Math.round(zoom * 100)}
          onChange={(e) => setZoom(Number(e.target.value) / 100)}
          className="w-32"
        />
        <span className="text-sm tabular-nums">{Math.round(zoom * 100)}%</span>
        <div className="h-4 w-px bg-border mx-1" />
        <button
          aria-label="Flip horizontal"
          className="p-1.5 rounded hover:bg-muted transition-colors cursor-pointer"
          onClick={toggleFlip}
        >
          <FlipHorizontal className="w-4 h-4" />
        </button>
        <button
          aria-label="Rotate 90° clockwise"
          className="p-1.5 rounded hover:bg-muted transition-colors cursor-pointer"
          onClick={rotate}
        >
          <RotateCw className="w-4 h-4" />
        </button>
        <button
          className="text-sm px-2 py-1 rounded hover:bg-muted transition-colors cursor-pointer"
          onClick={resetTo1to1}
        >
          1:1
        </button>
        <div className="flex-1" />
      </div>
      <div
        ref={containerRef}
        className="flex-1 overflow-hidden relative"
        style={{ cursor: dragging ? 'grabbing' : 'grab' }}
        {...dragHandlers}
      >
        {image.thumbnail_url && !loaded && (
          <img
            data-testid="viewer-img"
            src={image.thumbnail_url}
            alt=""
            aria-hidden="true"
            draggable={false}
            style={{
              position: 'absolute',
              left: '50%',
              top: '50%',
              transform,
              width: image.width ?? undefined,
              height: image.height ?? undefined,
              objectFit: 'contain',
              filter: 'blur(16px)',
            }}
          />
        )}
        {displaySrc && (
          <img
            src={displaySrc}
            alt={image.title}
            draggable={false}
            onLoad={() => setLoaded(true)}
            style={{
              position: 'absolute',
              left: '50%',
              top: '50%',
              transform,
              width: image.width ?? undefined,
              height: image.height ?? undefined,
              objectFit: 'contain',
              opacity: loaded ? 1 : 0,
            }}
          />
        )}
        {!loaded && (
          <div className="absolute inset-0 flex items-center justify-center">
            <Loader2 className="w-8 h-8 animate-spin text-white drop-shadow" />
          </div>
        )}
        <div className="absolute bottom-4 left-1/2 -translate-x-1/2 bg-black/50 text-white text-xs px-3 py-1 rounded-full whitespace-nowrap pointer-events-none">
          {`${image.title} · ${image.width} × ${image.height}`}
        </div>
      </div>
    </div>
  )
}

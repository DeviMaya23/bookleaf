import { useEffect, useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { Loader2, AlertCircle, ImageIcon } from 'lucide-react'
import { getSharedFolder } from '@/lib/share'
import MasonryLayout from '@/features/gallery/components/MasonryLayout'
import SharedImageCard from './SharedImageCard'
import SharedFolderPanel from './SharedFolderPanel'
import Lightbox from './Lightbox'
import { toGalleryImage } from '../lib/toGalleryImage'

export default function SharePage() {
  const { token = '' } = useParams<{ token: string }>()
  const [lightboxIndex, setLightboxIndex] = useState<number | null>(null)

  const [container, setContainer] = useState<HTMLDivElement | null>(null)
  const [containerWidth, setContainerWidth] = useState(660)

  useEffect(() => {
    if (!container) return
    const observer = new ResizeObserver(([entry]) => {
      setContainerWidth(entry.contentRect.width)
    })
    observer.observe(container)
    const rafId = requestAnimationFrame(() => {
      const w = container.getBoundingClientRect().width
      if (w > 0) setContainerWidth(w)
    })
    return () => { observer.disconnect(); cancelAnimationFrame(rafId) }
  }, [container])

  const { data, error, isLoading } = useQuery({
    queryKey: ['share', token],
    queryFn: () => getSharedFolder(token),
    retry: false,
  })

  if (isLoading) {
    return (
      <div className="h-screen flex items-center justify-center bg-background text-foreground">
        <Loader2 className="w-8 h-8 animate-spin text-muted-foreground" />
      </div>
    )
  }

  if (error || !data) {
    return (
      <div className="h-screen flex flex-col items-center justify-center gap-3 text-center px-4 bg-background text-foreground">
        <AlertCircle className="w-10 h-10 text-destructive" />
        <p className="text-sm text-muted-foreground">This link is invalid or has expired</p>
        <Link to="/" className="text-xs text-muted-foreground hover:text-foreground transition-colors">
          Shared via Bookleaf
        </Link>
      </div>
    )
  }

  const { folder, images } = data
  const adaptedImages = images.map((img, i) => toGalleryImage(img, i))

  return (
    <div className="h-screen flex flex-col bg-background text-foreground">
      <div className="flex-shrink-0 flex items-center h-12 px-4 gap-3 border-b">
        <Link to="/" className="font-serif text-base font-semibold hover:text-muted-foreground transition-colors">
          Bookleaf
        </Link>
        <div className="h-4 w-px bg-border" />
        <span className="text-sm text-muted-foreground truncate">{folder.name}</span>
      </div>

      <div className="flex-1 flex flex-col overflow-y-auto sm:flex-row sm:overflow-hidden">
        <div ref={setContainer} className="p-6 sm:flex-1 sm:overflow-y-auto">
          {images.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-20 text-muted-foreground gap-3">
              <ImageIcon className="w-10 h-10" />
              <p className="text-sm">No images here yet</p>
            </div>
          ) : (
            <MasonryLayout
              images={adaptedImages}
              containerWidth={containerWidth}
              dropIndicatorId={null}
              renderCard={(image, imgHeight) => {
                const index = Number(image.id)
                return (
                  <SharedImageCard
                    key={image.id}
                    image={image}
                    imgHeight={imgHeight}
                    downloadUrl={images[index].download_url}
                    onDoubleClick={() => setLightboxIndex(index)}
                  />
                )
              }}
            />
          )}
        </div>

        <SharedFolderPanel token={token} folder={folder} imageCount={images.length} className="order-first sm:order-last" />
      </div>

      {lightboxIndex !== null && (
        <Lightbox
          images={images}
          index={lightboxIndex}
          onClose={() => setLightboxIndex(null)}
          onNavigate={(dir) => {
            setLightboxIndex((current) => {
              if (current === null) return current
              const next = current + dir
              if (next < 0 || next >= images.length) return current
              return next
            })
          }}
        />
      )}
    </div>
  )
}

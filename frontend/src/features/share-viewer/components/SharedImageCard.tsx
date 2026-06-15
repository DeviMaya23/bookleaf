import { useState } from 'react'
import { Download, ImageIcon } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import type { Image } from '@/lib/images'

interface SharedImageCardProps {
  image: Image
  imgHeight: number
  downloadUrl: string
  onDoubleClick: () => void
}

export default function SharedImageCard({ image, imgHeight, downloadUrl, onDoubleClick }: SharedImageCardProps) {
  const [isSelected, setIsSelected] = useState(false)

  return (
    <div
      className={cn('rounded-lg overflow-hidden bg-card cursor-pointer', isSelected && 'ring-1 ring-primary')}
      onClick={() => setIsSelected((v) => !v)}
      onDoubleClick={onDoubleClick}
    >
      <div className="group relative bg-muted overflow-hidden rounded-t-lg" style={{ height: imgHeight }}>
        {image.thumbnail_url ? (
          <img
            src={image.thumbnail_url}
            alt={image.title}
            className="w-full h-full object-cover"
          />
        ) : (
          <div className="w-full h-full flex items-center justify-center">
            <ImageIcon className="w-8 h-8 text-muted-foreground" />
          </div>
        )}
        <Button
          variant="secondary"
          size="icon-sm"
          className="absolute bottom-2 right-2 rounded-full opacity-0 group-hover:opacity-100 transition-opacity"
          onClick={(e) => e.stopPropagation()}
          render={<a href={downloadUrl} download aria-label={`Download ${image.title}`} />}
        >
          <Download className="w-3.5 h-3.5" />
        </Button>
      </div>
      <div className="p-2">
        <p className="text-sm truncate">{image.title}</p>
      </div>
    </div>
  )
}

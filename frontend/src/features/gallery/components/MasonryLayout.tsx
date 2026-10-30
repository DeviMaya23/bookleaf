import { ImageIcon } from 'lucide-react'
import { computeMasonryLayout, GAP } from '../lib/masonry'

export interface MasonryImage {
  id: string
  title: string
  width: number | null
  height: number | null
  thumbnail_url: string | null
}

interface MasonryLayoutProps<T extends MasonryImage> {
  images: T[]
  containerWidth: number
  dropIndicatorId?: string | null
  renderCard: (image: T, imgHeight: number, isDropTarget: boolean) => React.ReactNode
}

export default function MasonryLayout<T extends MasonryImage>({ images, containerWidth, dropIndicatorId, renderCard }: MasonryLayoutProps<T>) {
  const { numCols, colWidth } = computeMasonryLayout(containerWidth)

  const columns: T[][] = Array.from({ length: numCols }, () => [])
  images.forEach((img, i) => {
    columns[i % numCols].push(img)
  })

  return (
    <div className="flex" style={{ gap: GAP }}>
      {columns.map((col, colIdx) => (
        <div key={colIdx} className="flex flex-col" style={{ width: colWidth, gap: GAP }}>
          {col.map((image) => {
            const ar = image.width && image.height ? image.width / image.height : 1
            const imgHeight = colWidth / ar
            return renderCard(image, imgHeight, dropIndicatorId === `image-${image.id}`)
          })}
        </div>
      ))}
    </div>
  )
}

interface MasonryCardContentProps {
  image: MasonryImage
  imgHeight: number
}

export function MasonryCardContent({ image, imgHeight }: MasonryCardContentProps) {
  return (
    <>
      <div className="bg-muted overflow-hidden rounded-t-lg" style={{ height: imgHeight }}>
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
      </div>
      <div className="p-2">
        <p className="text-sm truncate">{image.title}</p>
      </div>
    </>
  )
}

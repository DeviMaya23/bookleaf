import { useEffect, useRef, useState } from 'react'
import { Loader2, ImageIcon } from 'lucide-react'
import { SortableContext, useSortable } from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
import {
  ContextMenu,
  ContextMenuContent,
  ContextMenuItem,
  ContextMenuSeparator,
  ContextMenuTrigger,
} from '@/components/ui/context-menu'
import { Button } from '@/components/ui/button'
import type { Image } from '@/lib/images'
import type { AppView } from '@/lib/view'
import MasonryLayout, { MasonryCardContent } from './MasonryLayout'
import DeleteImageDialog from './DeleteImageDialog'
import { useImageLifecycle } from '../hooks/useImageLifecycle'
import { useManualReorder } from '../hooks/useManualReorder'
import { useGalleryImages, EMPTY_FILTER } from '../hooks/useGalleryImages'
import type { SortBy, SortDir } from '../hooks/useGalleryControls'


export type LayoutMode = 'masonry'

export interface SortEndTrigger {
  activeId: string
  overId: string
  ts: number
}

interface ImageCardProps {
  image: Image
  imgHeight: number
  isTrash: boolean
  isDropTarget: boolean
  currentFolderId: string | null
  onAction: (image: Image) => void
  onDeletePermanent?: (image: Image) => void
  onSelect: (image: Image) => void
  onDoubleClick?: (image: Image) => void
}

function ImageCard({ image, imgHeight, isTrash, isDropTarget, currentFolderId, onAction, onDeletePermanent, onSelect, onDoubleClick }: ImageCardProps) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({
    id: `image-${image.id}`,
    disabled: isTrash,
    data: { type: 'image', imageId: image.id, currentFolderId, thumbnailUrl: image.thumbnail_url },
  })

  const style: React.CSSProperties = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.4 : 1,
  }

  return (
    <ContextMenu>
      <ContextMenuTrigger>
        <div
          ref={setNodeRef}
          style={style}
          {...listeners}
          {...attributes}
          className={`cursor-pointer rounded-lg overflow-hidden bg-card${isDropTarget ? ' ring-2 ring-primary' : ''}`}
          onClick={() => onSelect(image)}
          onDoubleClick={() => onDoubleClick?.(image)}
        >
          <MasonryCardContent image={image} imgHeight={imgHeight} />
        </div>
      </ContextMenuTrigger>
      <ContextMenuContent>
        {isTrash ? (
          <>
            <ContextMenuItem onClick={() => onAction(image)}>Restore</ContextMenuItem>
            <ContextMenuSeparator />
            <ContextMenuItem
              onClick={() => onDeletePermanent?.(image)}
              className="text-destructive focus:text-destructive"
            >
              Delete permanently
            </ContextMenuItem>
          </>
        ) : (
          <ContextMenuItem
            onClick={() => onAction(image)}
            className="text-destructive focus:text-destructive"
          >
            Delete
          </ContextMenuItem>
        )}
      </ContextMenuContent>
    </ContextMenu>
  )
}

interface ImageGridProps {
  view: AppView
  layoutMode?: LayoutMode
  searchTerm: string
  debouncedSearchTerm: string
  sortBy: SortBy
  sortDir: SortDir | undefined
  filterTagIds?: string[]
  filterMimeTypes?: string[]
  filterFolderIds?: string[]
  onImageSelect: (image: Image) => void
  onImageDoubleClick?: (image: Image) => void
  onImageDeleted?: (id: string) => void
  sortEndTrigger?: SortEndTrigger | null
}

export default function ImageGrid({ view, layoutMode = 'masonry', searchTerm, debouncedSearchTerm, sortBy, sortDir, filterTagIds = EMPTY_FILTER, filterMimeTypes = EMPTY_FILTER, filterFolderIds = EMPTY_FILTER, onImageSelect, onImageDoubleClick, onImageDeleted, sortEndTrigger }: ImageGridProps) {
  const isTrash = view.type === 'trash'
  const isFolderView = view.type === 'folder'

  const containerRef = useRef<HTMLDivElement>(null)
  const [containerWidth, setContainerWidth] = useState(660)

  useEffect(() => {
    const el = containerRef.current
    if (!el) return
    const observer = new ResizeObserver(([entry]) => {
      setContainerWidth(entry.contentRect.width)
    })
    observer.observe(el)
    // Fallback for initial measurement when ResizeObserver fires before layout completes
    const rafId = requestAnimationFrame(() => {
      const w = el.getBoundingClientRect().width
      if (w > 0) setContainerWidth(w)
    })
    return () => { observer.disconnect(); cancelAnimationFrame(rafId) }
  }, [])

  const { images, isLoading, fetchNextPage, hasNextPage, isFetchingNextPage } =
    useGalleryImages({ view, searchTerm, debouncedSearchTerm, sortBy, sortDir, filterTagIds, filterMimeTypes, filterFolderIds })

  const { orderedImages, dragOverId, removeImage, sortableItems } =
    useManualReorder(view, sortBy, isFolderView, images, sortEndTrigger)

  const { handleAction, confirmDeleteImage, openDeleteConfirm, closeDeleteConfirm, confirmDelete } =
    useImageLifecycle(isTrash, removeImage, onImageDeleted)

  const folderId = view.type === 'folder' ? view.id : null

  if (layoutMode !== 'masonry') {
    console.warn(`ImageGrid: unsupported layoutMode "${layoutMode}"`)
    return null
  }

  const grid = (
    <MasonryLayout
      images={orderedImages}
      containerWidth={containerWidth}
      dropIndicatorId={dragOverId}
      renderCard={(image, imgHeight, isDropTarget) => (
        <ImageCard
          key={image.id}
          image={image}
          imgHeight={imgHeight}
          isTrash={isTrash}
          isDropTarget={isDropTarget}
          currentFolderId={folderId}
          onAction={handleAction}
          onDeletePermanent={openDeleteConfirm}
          onSelect={onImageSelect}
          onDoubleClick={onImageDoubleClick}
        />
      )}
    />
  )

  return (
    <div ref={containerRef} className="w-full">
      {isLoading ? (
        <div className="flex items-center justify-center py-20">
          <Loader2 className="w-8 h-8 animate-spin text-muted-foreground" />
        </div>
      ) : orderedImages.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-20 text-muted-foreground gap-3">
          <ImageIcon className="w-10 h-10" />
          <p className="text-sm">
            {searchTerm.trim()
              ? 'No images match your search'
              : isTrash ? 'Trash is empty' : 'No images here yet'}
          </p>
        </div>
      ) : isFolderView && sortBy === 'manual' ? (
        <SortableContext items={sortableItems} strategy={() => null}>
          {grid}
        </SortableContext>
      ) : (
        grid
      )}

      {hasNextPage && (
        <div className="flex justify-center mt-6">
          <Button
            variant="outline"
            onClick={() => fetchNextPage()}
            disabled={isFetchingNextPage}
          >
            {isFetchingNextPage ? (
              <>
                <Loader2 className="w-4 h-4 animate-spin mr-2" />
                Loading...
              </>
            ) : (
              'Load more'
            )}
          </Button>
        </div>
      )}

      <DeleteImageDialog
        image={confirmDeleteImage}
        onCancel={closeDeleteConfirm}
        onConfirm={confirmDelete}
      />
    </div>
  )
}
